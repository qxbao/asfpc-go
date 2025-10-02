import gc
import json
import logging
import pickle
from datetime import datetime
from pathlib import Path
from typing import Any

import numpy as np
import optuna
import pandas as pd
import xgboost as xgb
from sklearn.metrics import r2_score, root_mean_squared_error
from sklearn.model_selection import train_test_split
from xgboost.callback import LearningRateScheduler

from database.services.request import RequestService
from ml.model_utils import ModelUtility


class PotentialCustomerScoringModel:
  """Class for training, testing, and using a potential customer scoring model using XGBoost with Optuna hyperparameter tuning."""

  def __init__(self, request_id: int | None = None):
    self.request_id = request_id
    self.trials = 20
    self.model = None
    self.encoders = {}
    self.scaler = None
    self.embedding_dim = 768
    self.logger = logging.getLogger(__name__)
    self.use_gpu = True
    self.rs = RequestService()

  def load_data(self, df: pd.DataFrame) -> None:
    """
    Load and preprocess data for training and testing.

    Args:
        df (pd.DataFrame): Input DataFrame containing features (embedding, gender, locale, relationship_status) and labels (gemini_score).

    """
    label_col = self.util.label_column
    df = df.copy()
    df["score_bin"] = pd.qcut(df[label_col], q=5, duplicates="drop")

    try:
      train_df, test_df = train_test_split(
        df, test_size=0.2, random_state=42, stratify=df["score_bin"]
      )
      self.logger.info("Used stratified sampling for train/test split")
    except ValueError as e:
      self.logger.warning("Stratified sampling failed: %s. Using random split.", e)
      train_df, test_df = train_test_split(df, test_size=0.2, random_state=42)

    train_df = train_df.drop("score_bin", axis=1)
    test_df = test_df.drop("score_bin", axis=1)

    self.X_train, scaler, encoders = self.util.prepare_features(
      train_df, self.scaler, self.encoders
    )
    self.X_test, _, _ = self.util.prepare_features(test_df, scaler, encoders)

    self.y_train = train_df[label_col].to_numpy().astype(np.float32)
    self.y_test = test_df[label_col].to_numpy().astype(np.float32)
    self.scaler = scaler
    self.encoders = encoders

  def auto_tune(self) -> dict[str, Any]:  # noqa: PLR0911
    """
    Automatically tune hyperparameters using Optuna. GPU availability is automatically detected.

    Raises:
        ValueError: If data is not loaded.

    Returns:
        params: dict[str, Any]: Best hyperparameters found.

    """
    if not hasattr(self, "X_train"):
      msg = "Data not loaded. Call load_data first."
      raise ValueError(msg)
    if self.use_gpu:
      self.use_gpu = self._test_gpu()

    sample_size = min(self.util.get_sample_size(self.use_gpu), len(self.X_train))
    x_sample = self.X_train[:sample_size]
    y_sample = self.y_train[:sample_size]
    objective = self.util.get_optuna_objective(x_sample, y_sample, self.use_gpu)

    try:
      n_trials = self.trials
      timeout = 3600 if self.use_gpu else 7200
      study = optuna.create_study(
        direction="minimize",
        study_name=f"xgboost_tuning_{'gpu' if self.use_gpu else 'cpu'}_{datetime.now(tz=datetime.UTC).strftime('%Y%m%d_%H%M%S')}",
        sampler=optuna.samplers.TPESampler(
          seed=42,
          n_startup_trials=8 if self.use_gpu else 10,
          n_ei_candidates=16 if self.use_gpu else 24,
          multivariate=True,
          group=True,
        ),
        pruner=optuna.pruners.MedianPruner(
          n_startup_trials=3 if self.use_gpu else 5,
          n_warmup_steps=8 if self.use_gpu else 10,
          interval_steps=3 if self.use_gpu else 5,
        ),
      )

      self.logger.info(
        "Starting %s optimization with %d trials, timeout %ds",
        "GPU" if self.use_gpu else "CPU",
        n_trials,
        timeout,
      )
      ur_callback = self.util.get_ur_callback(self.request_id, self.trials)

      study.optimize(
        objective,
        n_trials=self.trials,
        timeout=timeout,
        show_progress_bar=False,
        gc_after_trial=self.use_gpu,
        callbacks=[ur_callback],
      )

      best_params = study.best_params
      best_value = study.best_value
      self.logger.info(
        "Optuna %s optimization completed. Best RMSE: %.4f",
        "GPU" if self.use_gpu else "CPU",
        best_value,
      )
      self.logger.info("Best parameters: %s", best_params)

      if best_value == float("inf"):
        self.logger.warning(
          "All %s trials failed (best RMSE is inf)", "GPU" if self.use_gpu else "CPU"
        )
        if self.use_gpu:
          self.logger.info("Switching to CPU optimization due to GPU failure")
          gc.collect()
          self.use_gpu = False
          return self.auto_tune()
        self.logger.error(
          "CPU trials also failed, returning empty parameters to use defaults"
        )
        return {}
      if self.use_gpu:
        gc.collect()
      if not best_params:
        self.logger.warning("Optuna returned empty parameters, using defaults")
        return {}
      try:
        return {
          "booster": best_params["booster"],
          "grow_policy": best_params["grow_policy"],
          "nthread": best_params["nthread"],
          "eta": best_params["eta"],
          "max_depth": best_params["max_depth"],
          "min_child_weight": best_params["min_child_weight"],
          "subsample": best_params["subsample"],
          "colsample_bytree": best_params["colsample_bytree"],
          "gamma": best_params["gamma"],
          "reg_alpha": best_params["reg_alpha"],
          "reg_lambda": best_params["reg_lambda"],
          "lr_decay": best_params["lr_decay"],
          "n_estimators": best_params["n_estimators"],
        }
      except KeyError:
        self.logger.exception("Missing expected parameter in best_params")
        return {}
    except Exception as e:
      self.logger.warning(
        "Optuna optimization failed with %s: %s", "GPU" if self.use_gpu else "CPU", e
      )
      if self.use_gpu:
        self.logger.info(
          "GPU optimization failed, falling back to CPU for hyperparameter optimization"
        )
        gc.collect()
        self.use_gpu = False
        return self.auto_tune()
      self.logger.exception("CPU optimization also failed, using default parameters")
      return {}

  def train(self, auto_tune: bool = True):
    dtrain = xgb.DMatrix(self.X_train, label=self.y_train, enable_categorical=False)
    dval = xgb.DMatrix(self.X_test, label=self.y_test, enable_categorical=False)

    params = self.util.get_base_params(self.use_gpu)
    params.update(self.util.recommended_params)
    self.train_params = params
    num_boost_round = 800
    lr_decay = 0.95

    if auto_tune:
      try:
        best_params = self.auto_tune()
        if best_params:
          if "n_estimators" in best_params:
            num_boost_round = best_params.pop("n_estimators")
          if "lr_decay" in best_params:
            lr_decay = best_params.pop("lr_decay")
          params.update(best_params)
          self.logger.info("Using Optuna-tuned parameters: %s", best_params)
      except KeyError as e:
        self.logger.warning(
          "Optuna tuning failed due to missing key, using default parameters: %s", e
        )

    callbacks = []
    if lr_decay is not None and "eta" in params:
      eta = params.get("eta", 0.05)
      lrdecay_callback = LearningRateScheduler(lambda epoch: eta * (lr_decay**epoch))
      callbacks.append(lrdecay_callback)
      self.logger.info("Using learning rate decay: %s (base eta: %s)", lr_decay, eta)
    elif lr_decay is not None:
      self.logger.warning(
        "Learning rate decay requested but 'eta' not found in params, skipping decay"
      )

    try:
      self.logger.info("Training model with %s", "GPU" if self.use_gpu else "CPU")
      self.model = self._execute_training(
        params, dtrain, dval, num_boost_round, callbacks
      )
      self.logger.info("Model training completed successfully")
    except Exception as e:
      if self.use_gpu:
        self.logger.warning("GPU training failed: %s", e)
        self.logger.info("Falling back to CPU training")
        self.use_gpu = False
        params.update(self.util.get_base_params(self.use_gpu))
        self.model = self._execute_training(
          params, dtrain, dval, num_boost_round, callbacks
        )
        self.logger.info("Model training completed successfully with CPU fallback")
      else:
        self.logger.exception("CPU training also failed")
        raise

  def test(self):
    if self.model is None:
      msg = "No model to test"
      raise ValueError(msg)

    if not hasattr(self, "X_test") or not hasattr(self, "y_test"):
      msg = "Test data not loaded"
      raise ValueError(msg)

    dtest = xgb.DMatrix(self.X_test)
    y_pred = self.model.predict(dtest)
    rmse = root_mean_squared_error(self.y_test, y_pred)

    self.test_results = {
      "rmse": rmse,
      "r2": r2_score(self.y_test, y_pred),
      "mae": np.mean(np.abs(self.y_test - y_pred)),
    }
    self.test_results = self.util.convert_numpy_types(self.test_results)
    return self.test_results

  def load_model(self, model_name: str):
    self.util = ModelUtility(model_name)
    model_dir = self.util.model_path
    self.model = xgb.Booster()
    try:
      self.model.load_model(str(model_dir / "model.json"))
    except FileNotFoundError as err:
      msg = "There's no model match this model name."
      raise FileNotFoundError(msg) from err
    encoders_path = model_dir / "encoders.pkl"
    if Path(encoders_path).exists():
      with Path.open(encoders_path, "rb") as f:
        self.encoders = pickle.load(f)  # noqa: S301
      self.logger.info("Encoders loaded from %s", encoders_path)
    else:
      self.logger.warning(
        "Encoders file not found at %s. Model loaded without encoders.", encoders_path
      )
      self.encoders = {}

    scaler_path = model_dir / "scalers.pkl"
    if Path(scaler_path).exists():
      with Path.open(scaler_path, "rb") as f:
        self.scaler = pickle.load(f)  # noqa: S301
      self.logger.info("Scaler loaded from %s", scaler_path)
    else:
      self.logger.warning(
        "Scaler file not found at %s. Model loaded without scaler.", scaler_path
      )
      self.scaler = None
    self.logger.info("Model loaded from %s", model_dir)

  def predict(self, df: pd.DataFrame) -> Any:
    if self.model is None:
      err_msg = "No model found"
      raise ValueError(err_msg)
    x, scaler, encoders = self.util.prepare_features(df, self.scaler, self.encoders)
    self.scaler = scaler
    self.encoders = encoders
    dmatrix = xgb.DMatrix(x)
    predictions = self.model.predict(dmatrix)
    return self.util.convert_numpy_types(predictions)

  def save_model(self):
    if self.model is None:
      err_msg = "No model to save"
      raise ValueError(err_msg)
    model_dir = self.util.model_path
    Path.mkdir(model_dir, exist_ok=True)

    self.model.save_model(str(model_dir / "model.json"))

    with Path.open(model_dir / "encoders.pkl", "wb") as f:
      pickle.dump(self.encoders, f)

    if self.scaler is not None:
      with Path.open(model_dir / "scalers.pkl", "wb") as f:
        pickle.dump(self.scaler, f)

    if not hasattr(self, "test_results"):
      self.test_results = {}
    self.test_results["saved_at"] = datetime.now(tz=datetime.UTC).isoformat()
    self.test_results["is_gpu"] = self.use_gpu
    self.test_results["train_params"] = (
      self.train_params if hasattr(self, "train_params") else {}
    )
    metadata = self.util.convert_numpy_types(self.test_results)
    with Path.open(model_dir / "metadata.json", "w") as f:
      json.dump(metadata, f, indent=2)
    self.logger.info("Model saved to %s", model_dir)

  def _execute_training(
    self,
    params: dict,
    dtrain: xgb.DMatrix,
    dval: xgb.DMatrix,
    num_boost_round: int,
    callbacks: list | None = None,
  ) -> xgb.Booster:
    """Helper method to execute XGBoost training with given parameters"""
    return xgb.train(
      params,
      dtrain,
      num_boost_round=num_boost_round,
      evals=[(dtrain, "train"), (dval, "eval")],
      early_stopping_rounds=20,
      verbose_eval=False,
      callbacks=callbacks if callbacks else None,
    )

  def _test_gpu(self) -> bool:
    if not hasattr(self, "X_train") or not hasattr(self, "y_train"):
      err_msg = "Data not loaded. Call load_data() first."
      raise ValueError(err_msg)
    try:
      test_matrix = xgb.DMatrix(self.X_train[:100], label=self.y_train[:100])
      test_params = {"device": "cuda", "tree_method": "hist", "max_bin": 256}
      xgb.train(test_params, test_matrix, num_boost_round=1, verbose_eval=False)
      self.logger.info("GPU is available and working")
      del test_matrix
      gc.collect()
    except xgb.core.XGBoostError as e:
      self.logger.warning("GPU test failed: %s", e)
      self.logger.info("Switching to CPU for optimization")
      self.use_gpu = False
    else:
      return True
