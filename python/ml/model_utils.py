import asyncio
import datetime
import gc
import logging
import threading
from collections.abc import Sequence
from pathlib import Path
from typing import Any, TypeVar

import numpy as np
import optuna
import pandas as pd
import xgboost as xgb
from dateutil.parser import parse
from optuna.study.study import ObjectiveFuncType
from optuna.trial import Trial
from sklearn.calibration import LabelEncoder
from sklearn.discriminant_analysis import StandardScaler
from xgboost.callback import LearningRateScheduler

from database.database import Database
from database.services.request import RequestService

T = TypeVar("T")

class ModelUtility:
  MAX_AGE = 120
  def __init__(self, model_name):
    self.model_name = model_name
    self.logger = logging.getLogger("ModelUtility")

  @property
  def model_path(self) -> Path:
    return Path.cwd() / "resources" / "models" / self.model_name

  @property
  def required_features(self) -> list[str]:
    return ["embedding", "gender", "locale", "relationship_status"]

  @property
  def label_column(self) -> str:
    return "gemini_score"

  @property
  def embedding_dimension(self) -> int:
    return 768

  @property
  def recommended_params(self) -> dict:
    return {
        "booster": "gbtree",            # Booster ổn định nhất cho regression
        "grow_policy": "lossguide",     # Tốt cho dữ liệu cao chiều (embedding)
        "eta": 0.03,                    # Giảm learning rate để hội tụ mượt hơn
        "max_depth": 5,                 # Depth nhỏ hơn để tránh overfit
        "max_leaves": 64,               # Giới hạn số leaf để kiểm soát complexity
        "min_child_weight": 4,          # Tăng để tránh split nhỏ vô nghĩa
        "subsample": 0.8,               # Random row sampling
        "colsample_bytree": 0.7,        # Random feature sampling
        "colsample_bylevel": 0.8,
        "gamma": 0.15,                  # Pruning nhẹ, ổn định
        "reg_alpha": 0.6,               # Mạnh tay với L1 để loại bỏ noise
        "reg_lambda": 1.5,              # L2 để ổn định weight
        "nthread": 4,
        "lr_decay": 0.95,               # Cho phép decay nhẹ dần learning rate
        "verbosity": 0
    }

  def get_base_params(self, use_gpu: bool) -> dict:
    base = {
      "objective": "reg:squarederror",
      "eval_metric": "rmse",
      "seed": 42,
    }
    if use_gpu:
      # Recommended settings for GPU training (XGBoost)
      # Stop using device = gpu, tree_method = gpu_hist as it's deprecated
      base["device"] = "cuda"
      base["tree_method"] = "hist"
      base["max_bin"] = 256  # Use smaller max_bin for faster training on GPU
      self.logger.info(
        "Using GPU-optimized parameters (device=cuda, tree_method=hist, max_bin=256)"
      )
    else:
      base["device"] = "cpu"
      base["tree_method"] = "hist"
      self.logger.info("Using CPU parameters (device=cpu, tree_method=hist)")
    return base

  def get_sample_size(self, use_gpu: bool) -> int:
    return 12000 if use_gpu else 10000

  def get_ur_callback(
    self, request_id: int | None, total_trials: int
  ) -> "UpdateRequestCallback":
    """Create a callback to update request status during Optuna trials."""
    return UpdateRequestCallback(request_id, self.logger, total_trials)

  def get_optuna_objective(
    self, x: np.ndarray, y: np.ndarray, use_gpu: bool
  ) -> ObjectiveFuncType:
    """Create an Optuna objective function for hyperparameter tuning."""
    base_params = self.get_base_params(use_gpu)

    def objective(trial: Trial) -> float | Sequence[float]:
      try:
        params = base_params.copy()
        grow_policy = trial.suggest_categorical("grow_policy", ["depthwise", "lossguide"])
        params.update(
          {
            "booster": trial.suggest_categorical("booster", ["gbtree", "dart"]),
            "grow_policy": grow_policy,
            "verbosity": 0,
            "nthread": trial.suggest_int("nthread", 1, 8),
            "eta": trial.suggest_float("eta", 0.01, 0.07, log=True),
            "max_depth": trial.suggest_int("max_depth", 3, 6),
            "max_leaves": trial.suggest_int("max_leaves", 16, 256) if grow_policy == "lossguide" else 0,
            "min_child_weight": trial.suggest_float("min_child_weight", 2, 8),
            "subsample": trial.suggest_float("subsample", 0.7, 0.9),
            "colsample_bytree": trial.suggest_float("colsample_bytree", 0.6, 0.85),
            "colsample_bylevel": trial.suggest_float("colsample_bylevel", 0.6, 1.0),
            "colsample_bynode": trial.suggest_float("colsample_bynode", 0.6, 1.0),
            "gamma": trial.suggest_float("gamma", 0, 0.3),
            "reg_alpha": trial.suggest_float("reg_alpha", 0.2, 1.0),
            "reg_lambda": trial.suggest_float("reg_lambda", 1.0, 2.0),
            "lr_decay": trial.suggest_float("lr_decay", 0.9, 1.0),
          }
        )
        self.logger.info("Trial parameters: %s", params)
        n_estimators = trial.suggest_int("n_estimators", 100, 500)
        dtrain = xgb.DMatrix(x, label=y, enable_categorical=False)
        lrdecay_callback = LearningRateScheduler(
          lambda epoch: params["eta"] * (params["lr_decay"] ** epoch)
        )

        cv_results = xgb.cv(
          params,
          dtrain,
          num_boost_round=n_estimators,
          nfold=5,
          metrics=["rmse"],
          early_stopping_rounds=20,
          seed=42,
          shuffle=True,
          callbacks=[lrdecay_callback],
          verbose_eval=False,
        )

        rmse_values = cv_results["test-rmse-mean"]
        if isinstance(rmse_values, pd.Series):
          best_rmse = rmse_values.min()
        else:
          best_rmse = float(rmse_values)
        del dtrain
        gc.collect()
      except Exception as e:  # noqa: BLE001
        self.logger.warning(
          "Trial failed with %s: %s", "GPU" if use_gpu else "CPU", e
        )
        if use_gpu:
          gc.collect()
        return float("inf")
      else:
        return best_rmse

    return objective

  def get_age(self, birthday: str) -> int | float:
    try:
      birth_date = parse(birthday, dayfirst=False)
      today = datetime.datetime.now(tz=datetime.UTC).date()
      age = (today - birth_date.date()).days / 365.25
      if not (0 <= age <= self.MAX_AGE):
        self.logger.warning("Unrealistic age %d from birthday %s", age, birthday)
        return np.nan
    except (ValueError, TypeError):
      return np.nan
    else:
      return age

  def prepare_features(
    self, df: pd.DataFrame, scaler: StandardScaler | None, encoders: dict | None
  ) -> tuple[np.ndarray, StandardScaler, dict[str, LabelEncoder]]:
    """Prepare feature matrix X from DataFrame, applying scaling and encoding as needed, and returning scalers/encoders if they were missing in the input."""
    for feature in self.required_features:
      if feature not in df.columns:
        err_msg = f"Missing required feature: {feature}"
        raise ValueError(err_msg)
    x_emb = np.vstack(
      [self.validate_embedding(emb) for emb in df["embedding"].to_numpy()]
    )
    if not encoders:
      encoders = {}
    if not scaler:
      scaler = StandardScaler()
      x_emb = scaler.fit_transform(x_emb)
    else:
      x_emb = scaler.transform(x_emb)
    cate_features = ["gender", "locale", "relationship_status"]
    x_cate = []
    for col in cate_features:
      # Use frequency-based encoding for high-cardinality features
      min_freq = 0.01 if col == "locale" else 0.005  # Stricter for locale
      encoded_values = self.encode_with_frequency(df, col, encoders, min_freq=min_freq)
      x_cate.append(encoded_values)
    x_cate = np.vstack(x_cate).T
    # Keep NaN for missing age values - XGBoost handles them natively
    x_age = df["birthday"]\
        .fillna("")\
        .infer_objects(copy=False)\
        .apply(self.get_age)\
        .to_numpy()\
        .reshape(-1, 1)
    # Don't convert NaN to -1, XGBoost will treat NaN as missing value indicator
    x = np.hstack([x_emb, x_cate.astype(np.float32), x_age.astype(np.float32)])
    return x.astype(np.float32), scaler, encoders

  def encode_with_frequency(self, df: pd.DataFrame, col: str, encoders: dict[str, LabelEncoder], min_freq=0.01) -> dict[str, LabelEncoder]:
    """Group rare categories as 'Other'"""
    filled = df[col].fillna("(null)").astype(str)
    if col not in encoders:
        value_counts = filled.value_counts(normalize=True)
        rare_categories = value_counts[value_counts < min_freq].index

        filled = filled.replace(rare_categories, "Other")
        encoders[col] = LabelEncoder()
        encoders[col].fit(filled)
    else:
        # Handle unseen categories
        known_categories = set(encoders[col].classes_)
        filled = filled.apply(lambda x: x if x in known_categories else "Other")
    return encoders[col].transform(filled)

  def validate_embedding(self, embedding: list[float]) -> np.ndarray:
    try:
        arr = np.array(embedding, dtype=np.float32).flatten()

        # Check dimension
        if len(arr) != self.embedding_dimension:
            self.logger.warning("Expected %dD embedding, got %dD", self.embedding_dimension, len(arr))
            # Pad or truncate
            if len(arr) < self.embedding_dimension:
                arr = np.pad(arr, (0, self.embedding_dimension - len(arr)))
            else:
                arr = arr[:self.embedding_dimension]

        # Check for invalid values
        if np.any(np.isnan(arr)) or np.any(np.isinf(arr)):
            self.logger.warning("Embedding contains NaN/Inf values")
            # Use mean embedding instead of zeros
            return self._get_mean_embedding()  # Store this from training data

        # Check if all zeros (likely missing)
        if np.allclose(arr, 0):
            self.logger.warning("All-zero embedding detected")
            return self._get_mean_embedding()

    except Exception:
        self.logger.exception("Failed to validate embedding")
        return self._get_mean_embedding()
    else:
        return arr

  def convert_numpy_types(
    self, obj: T
  ) -> Any:
    """Recursively convert numpy types to Python native types for JSON serialization"""
    if isinstance(obj, dict):
      result = {key: self.convert_numpy_types(value) for key, value in obj.items()}
    elif isinstance(obj, list):
      result = [self.convert_numpy_types(item) for item in obj]
    elif isinstance(obj, np.integer):
      result = int(obj)
    elif isinstance(obj, np.floating):
      result = float(obj)
    elif isinstance(obj, np.ndarray):
      result = obj.tolist()
    elif isinstance(obj, (np.bool_, bool)):
      result = bool(obj)
    else:
      result = obj
    return result

class UpdateRequestCallback:
  def __init__(
    self,
    request_id: int | None,
    logger: logging.Logger,
    total_trials: int,
  ):
    self.request_id = request_id
    self.rs = RequestService()
    self.trial_count = 0
    self.logger = logger
    self.total_trials = total_trials

  def __call__(self, _: optuna.Study, __: optuna.trial.FrozenTrial):
    self.trial_count += 1
    if self.request_id is not None:
      progress = min(0.9, self.trial_count / self.total_trials * 0.8 + 0.1)
      description = f"Optuna trial in progress: {self.trial_count}/{self.total_trials}"
      self.logger.info(
        "Updating request %s, Progress: %.2f%%", self.request_id, progress * 100
      )
    else:
      return
    thread = threading.Thread(
      target=self.isolated_update, args=(progress, description), daemon=True
    )
    thread.start()
    self.logger.debug("Started isolated update thread for request %s", self.request_id)

  def isolated_update(self, progress: float, description: str):
    try:
      loop = asyncio.new_event_loop()
      asyncio.set_event_loop(loop)

      try:

        async def update_request_operation(
          request_id: int, progress: float, description: str
        ):
          """Database operation to update request status"""
          rs = RequestService()
          db_session = Database.get_isolated_session()
          return await rs.update_request(
            request_id,
            session=db_session,
            progress=progress,
            description=description,
            status=1,
          )

        if self.request_id is not None:
          success = loop.run_until_complete(
            update_request_operation(self.request_id, progress, description)
          )
          if success:
            self.logger.info("Successfully updated request %s", self.request_id)
          else:
            self.logger.warning("Request %s not found", self.request_id)
        else:
          self.logger.debug("No request_id provided, skipping update")
      finally:
        loop.close()
    except RuntimeError:
      self.logger.exception("Failed to update request status")
