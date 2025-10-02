import asyncio
import datetime
import gc
import logging
import threading
from collections.abc import Sequence
from pathlib import Path
from typing import Any

import numpy as np
import optuna
import pandas as pd
import xgboost as xgb
from optuna.study.study import ObjectiveFuncType, Trial
from sklearn.calibration import LabelEncoder
from sklearn.discriminant_analysis import StandardScaler
from xgboost.callback import LearningRateScheduler

from database.database import Database
from database.services.request import RequestService


class ModelUtility:
  def __init__(self, model_name):
    self.model_name = model_name
    self.logger = logging.getLogger("ModelUtility")

  @property
  def model_path(self) -> str:
    return str(Path.cwd() / "resources" / "models" / self.model_name)

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
      "booster": "gbtree",  # Standard gradient boosting
      "grow_policy": "lossguide",  # Loss-guided growth for better accuracy
      "eta": 0.05,  # Lower learning rate for smoother convergence
      "max_depth": 5,  # Moderate depth to prevent overfitting on small score range
      "min_child_weight": 3,  # Higher to reduce noise in probability predictions
      "subsample": 0.85,  # High subsample for stability
      "colsample_bytree": 0.85,  # High feature sampling for robustness
      "gamma": 0.1,  # Small regularization for pruning
      "reg_alpha": 0.3,  # L1 regularization to handle sparse features (embeddings)
      "reg_lambda": 1.2,  # L2 regularization for smooth predictions
      "nthread": 4,  # Parallel threads
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
    return UpdateRequestCallback(request_id, self.rs, self.logger, total_trials)

  def get_optuna_objective(
    self, x: np.ndarray, y: np.ndarray, use_gpu: bool
  ) -> ObjectiveFuncType:
    """Create an Optuna objective function for hyperparameter tuning."""
    base_params = self.get_base_params(use_gpu)

    def objective(trial: Trial) -> float | Sequence[float]:
      try:
        params = base_params.copy()
        params.update(
          {
            "booster": trial.suggest_categorical("booster", ["gbtree", "dart"]),
            "grow_policy": trial.suggest_categorical(
              "grow_policy", ["depthwise", "lossguide"]
            ),
            "verbosity": 0,
            "nthread": trial.suggest_int("nthread", 1, 8),
            "eta": trial.suggest_float("eta", 0.03, 0.2, log=True),
            "max_depth": trial.suggest_int("max_depth", 4, 9),
            "min_child_weight": trial.suggest_int("min_child_weight", 1, 6),
            "subsample": trial.suggest_float("subsample", 0.7, 1.0),
            "colsample_bytree": trial.suggest_float("colsample_bytree", 0.7, 1.0),
            "gamma": trial.suggest_float("gamma", 0, 0.3),
            "reg_alpha": trial.suggest_float("reg_alpha", 0, 1.0),
            "reg_lambda": trial.suggest_float("reg_lambda", 0.8, 2.0),
            "lr_decay": trial.suggest_float("lr_decay", 0.8, 1),
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
          "Trial failed with %s: %s", "GPU" if self.use_gpu else "CPU", e
        )
        if self.use_gpu:
          gc.collect()
        return float("inf")
      else:
        return best_rmse

    return objective

  def get_age(self, birthday: str) -> int | float:
    """Convert birthday string in MM/DD/YYYY format to age in years. Returns NaN if invalid."""
    birthday_part_n = 3
    try:
      parts = birthday.split("/")
      if len(parts) == birthday_part_n:
        current_year = datetime.datetime.now(tz=datetime.UTC).date().year
        return current_year - int(parts[2])
    except ValueError:
      return np.nan

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
    if not scaler:
      scaler = StandardScaler()
      x_emb = scaler.fit_transform(x_emb)
    else:
      x_emb = scaler.transform(x_emb)
    cate_features = ["gender", "locale", "relationship_status"]
    x_cate = []
    for col in cate_features:
      filled = df[col].fillna("(null)").astype(str)
      if col not in encoders:
        encoders[col] = LabelEncoder()
        x_cate.append(encoders[col].fit_transform(filled))
      else:
        unseen_mask = ~filled.isin(self.encoders[col].classes_)
        if unseen_mask.any():
          self.logger.warning(
            "Found unseen labels in column '%s': %s", col, filled[unseen_mask].unique()
          )
          default_label = (
            "(null)"
            if "(null)" in encoders[col].classes_
            else encoders[col].classes_[0]
          )
          filled = filled.where(~unseen_mask, default_label)
        x_cate.append(encoders[col].transform(filled))
    x_cate = np.vstack(x_cate).T
    x_age = np.array(
      df["birthday"]
      .fillna("")
      .infer_objects(copy=False)
      .apply(self.get_age)
      .fillna(-1)
      .values
    ).reshape(-1, 1)
    x = np.hstack([x_emb, x_cate.astype(np.float32), x_age.astype(np.float32)])
    return x.astype(np.float32), scaler, encoders

  def validate_embedding(self, embedding: list[float]) -> np.ndarray:
    """Ensure embedding is a valid 1D numpy array of the correct dimension"""
    try:
      arr = np.array(embedding, dtype=np.float32)
      if arr.ndim == 1:
        return arr
      if arr.ndim == 2:  # noqa: PLR2004
        return arr.flatten()
      err_msg = "Embedding must be a 1D or 2D array"
      raise ValueError(err_msg)  # noqa: TRY301
    except ValueError:
      self.logger.warning("Invalid embedding found, replaced with zeros.")
      return np.zeros(self.embedding_dimension, dtype=np.float32)

  def convert_numpy_types(
    self, obj: dict | list | np.ndarray | np.integer | np.floating | np.bool_
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
    rs: RequestService,
    logger: logging.Logger,
    total_trials: int,
  ):
    self.request_id = request_id
    self.rs = rs
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
    except RuntimeError as e:
      self.logger.warning("Failed to update request status: %s", e)
