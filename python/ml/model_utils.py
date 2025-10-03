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
from dateutil.parser import parse
from optuna.study.study import ObjectiveFuncType
from optuna.trial import Trial
from sklearn.calibration import LabelEncoder
from sklearn.discriminant_analysis import StandardScaler
from sklearn.metrics import r2_score, root_mean_squared_error
from xgboost.callback import LearningRateScheduler

from database.database import Database
from database.services.request import RequestService


class ModelUtility:
  MAX_AGE = 120
  SMALL_DATASET_THRESHOLD = 2000
  MEDIUM_DATASET_THRESHOLD = 5000
  LARGE_DATASET_THRESHOLD = 10000

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

  def get_recommended_params(self, dataset_size: int) -> dict:
    """
    Get recommended hyperparameters based on dataset size.

    Args:
        dataset_size: Number of training samples

    Returns:
        Dictionary of recommended hyperparameters

    Strategy:
        - Small dataset (<1000): High regularization, simple model
        - Medium dataset (1000-5000): Balanced complexity
        - Large dataset (5000-10000): Moderate regularization
        - Very large dataset (>10000): Lower regularization, more complex model

    """
    # Small dataset: Prevent overfitting aggressively
    if dataset_size < self.SMALL_DATASET_THRESHOLD:
      self.logger.info("Small dataset (%d samples): Using high regularization", dataset_size)
      return {
          "booster": "gbtree",
          "grow_policy": "depthwise",     # Simpler growth for small data
          "eta": 0.05,                    # Higher LR for faster convergence
          "max_depth": 3,                 # Shallow trees to prevent overfit
          "max_leaves": 32,               # Fewer leaves
          "min_child_weight": 5,          # Strong constraint on splits
          "subsample": 0.7,               # Aggressive row sampling
          "colsample_bytree": 0.7,        # Aggressive feature sampling
          "colsample_bylevel": 0.8,
          "gamma": 0.25,                   # Strong pruning
          "reg_alpha": 1.0,               # Heavy L1 regularization
          "reg_lambda": 2.0,              # Heavy L2 regularization
          "nthread": 4,
          "lr_decay": 0.9,
          "verbosity": 0,
          "nfold": 3
      }

    # Medium dataset: Balanced approach
    if dataset_size < self.MEDIUM_DATASET_THRESHOLD:
      self.logger.info("Medium dataset (%d samples): Using balanced parameters", dataset_size)
      return {
          "booster": "gbtree",
          "grow_policy": "lossguide",
          "eta": 0.03,
          "max_depth": 4,                 # Moderate depth
          "max_leaves": 48,               # Moderate complexity
          "min_child_weight": 6,
          "subsample": 0.75,
          "colsample_bytree": 0.7,
          "colsample_bylevel": 0.75,
          "gamma": 0.2,
          "reg_alpha": 0.8,
          "reg_lambda": 2.0,
          "nthread": 4,
          "lr_decay": 0.94,
          "verbosity": 0,
          "nfold": 4
      }

    # Large dataset: Standard regularization
    if dataset_size < self.LARGE_DATASET_THRESHOLD:
      self.logger.info("Large dataset (%d samples): Using standard regularization", dataset_size)
      return {
          "booster": "gbtree",
          "grow_policy": "lossguide",
          "eta": 0.03,
          "max_depth": 5,
          "max_leaves": 64,
          "min_child_weight": 4,
          "subsample": 0.8,
          "colsample_bytree": 0.7,
          "colsample_bylevel": 0.8,
          "gamma": 0.15,
          "reg_alpha": 0.6,
          "reg_lambda": 1.5,
          "nthread": 4,
          "lr_decay": 0.96,
          "verbosity": 0,
          "nfold": 5
      }

    # Very large dataset: Allow more model complexity
    self.logger.info("Very large dataset (%d samples): Using lower regularization", dataset_size)
    return {
        "booster": "gbtree",
        "grow_policy": "lossguide",
        "eta": 0.02,                    # Lower LR for fine-tuning
        "max_depth": 6,                 # Deeper trees possible
        "max_leaves": 96,               # More leaves for complex patterns
        "min_child_weight": 3,          # Allow smaller splits
        "subsample": 0.85,              # Less aggressive sampling
        "colsample_bytree": 0.75,
        "colsample_bylevel": 0.85,
        "gamma": 0.1,                   # Lighter pruning
        "reg_alpha": 0.4,               # Lower L1 regularization
        "reg_lambda": 1.0,              # Lower L2 regularization
        "nthread": 4,
        "lr_decay": 0.98,
        "verbosity": 0,
        "nfold": 5
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

  def get_num_boost_round(self, dataset_size: int) -> int:
    if dataset_size < self.SMALL_DATASET_THRESHOLD:
      return 300
    if dataset_size < self.MEDIUM_DATASET_THRESHOLD:
      return 600
    if dataset_size < self.LARGE_DATASET_THRESHOLD:
      return 800
    return 1200

  def get_early_stopping_rounds(self, dataset_size: int) -> int:
    if dataset_size < self.SMALL_DATASET_THRESHOLD:
      return 15
    if dataset_size < self.MEDIUM_DATASET_THRESHOLD:
      return 25
    if dataset_size < self.LARGE_DATASET_THRESHOLD:
      return 35
    return 40

  def get_max_bin(self, dataset_size: int, use_gpu: bool) -> int:
    # GPU: Use fixed 256 for optimal CUDA memory management
    if use_gpu:
      return 256
    # CPU: Scale based on dataset size
    if dataset_size < self.SMALL_DATASET_THRESHOLD:
      return 128  # Small dataset: Fast training, less granularity needed
    if dataset_size < self.MEDIUM_DATASET_THRESHOLD:
      return 256  # Medium dataset: Balanced speed and accuracy
    if dataset_size < self.LARGE_DATASET_THRESHOLD:
      return 384  # Large dataset: More granularity
    return 512  # Very large dataset: Maximum granularity for complex patterns

  def get_nfold(self, dataset_size: int) -> int:
    if dataset_size < self.SMALL_DATASET_THRESHOLD:
      return 3
    if dataset_size < self.MEDIUM_DATASET_THRESHOLD:
      return 4
    return 5

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
          nfold=self.get_nfold(len(y)),
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

  def encode_with_frequency(self, df: pd.DataFrame, col: str, encoders: dict[str, LabelEncoder], min_freq=0.01):
    """
    Group rare categories as 'Other' during training.
    Handle unseen categories during prediction by mapping to most frequent category.

    Args:
        df: Input dataframe
        col: Column name to encode
        encoders: Dictionary of existing encoders
        min_freq: Minimum frequency threshold for rare categories

    Returns:
        Encoded values as numpy array

    """
    filled = df[col].fillna("(null)").astype(str)

    if col not in encoders:
        # Training phase: Group rare categories
        value_counts = filled.value_counts(normalize=True)
        rare_categories = value_counts[value_counts < min_freq].index

        if len(rare_categories) > 0:
            filled = filled.replace(rare_categories.tolist(), "Other")
            self.logger.info(
                "Column '%s': Grouped %d rare categories into 'Other' (min_freq=%.3f)",
                col, len(rare_categories), min_freq
            )

        encoders[col] = LabelEncoder()
        encoders[col].fit(filled)
        self.logger.info(
            "Column '%s': Fitted encoder with %d classes",
            col, len(encoders[col].classes_)
        )
    else:
        # Prediction phase: Handle unseen categories
        known_categories = set(encoders[col].classes_)
        unseen_mask = ~filled.isin(known_categories)

        if unseen_mask.any():
            unseen_count = unseen_mask.sum()

            # Strategy: Map unseen to most frequent class in training data
            # This is safer than creating "Other" which may not exist
            most_frequent_class = encoders[col].classes_[0]  # Classes are sorted by frequency
            filled = filled.copy()
            filled[unseen_mask] = most_frequent_class

            self.logger.warning(
                "Column '%s': Found %d unseen categories, mapped to '%s'",
                col, unseen_count, most_frequent_class
            )

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

  def _get_mean_embedding(self) -> np.ndarray:
    return np.full((self.embedding_dimension,), 0.01, dtype=np.float32)

  def convert_numpy_types(
    self, obj: object
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

  def calculate_test_results(
    self,
    y_test: np.ndarray,
    y_pred: np.ndarray,
    model: xgb.Booster
  ) -> dict[str, Any]:
    """
    Calculate comprehensive test results including metrics, predictions stats,
    residual analysis, and feature importance.

    Args:
        y_test: True target values
        y_pred: Predicted values
        model: Trained XGBoost model

    Returns:
        Dictionary containing all test metrics and statistics

    """
    self.logger.info(
        "Target value stats - min: %.6f, max: %.6f, mean: %.6f, zeros: %d/%d",
        y_test.min(), y_test.max(), y_test.mean(),
        np.sum(y_test == 0), len(y_test)
    )

    epsilon = 1e-10

    # RMSLE calculation (handles zeros by using log1p)
    y_test_pos = np.maximum(y_test, 0)
    y_pred_pos = np.maximum(y_pred, 0)
    rmsle = float(np.sqrt(np.mean((np.log1p(y_pred_pos) - np.log1p(y_test_pos)) ** 2)))

    # SMAPE calculation (symmetric, handles zeros with epsilon)
    smape = float(
      np.mean(2.0 * np.abs(y_pred - y_test) / (np.abs(y_test) + np.abs(y_pred) + epsilon)) * 100
    )

    # Basic metrics
    test_results: dict[str, Any] = {
        "rmse": float(root_mean_squared_error(y_test, y_pred)),
        "r2": float(r2_score(y_test, y_pred)),
        "mae": float(np.mean(np.abs(y_test - y_pred))),
        "rmsle": rmsle,
        "smape": smape,
    }

    # Prediction statistics
    test_results["prediction_stats"] = {
        "min": float(y_pred.min()),
        "max": float(y_pred.max()),
        "mean": float(y_pred.mean()),
        "std": float(y_pred.std()),
    }

    # Residual analysis
    residuals = y_test - y_pred
    test_results["residual_stats"] = {
        "mean": float(residuals.mean()),  # Should be ~0
        "std": float(residuals.std()),
        "bias_low_scores": float(residuals[y_test < y_test.mean()].mean()),
        "bias_high_scores": float(residuals[y_test >= y_test.mean()].mean()),
    }

    # Feature importance (top 10)
    importance = model.get_score(importance_type="gain")
    test_results["top_features"] = dict(
      sorted(importance.items(), key=lambda x: x[1], reverse=True)[:10]
    )

    self.logger.info("Test Results: %s", test_results)
    return test_results


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
