from datetime import datetime
import json
import logging
import os
import pandas as pd
from sklearn.preprocessing import LabelEncoder, StandardScaler
from sklearn.metrics import root_mean_squared_error, r2_score
from sklearn.model_selection import train_test_split
import xgboost as xgb
import numpy as np
import optuna
from optuna.integration import XGBoostPruningCallback

class PotentialCustomerScoringModel:
    model_path = os.path.join(os.getcwd(), "resources", "models")
    required_features = [
        "embedding",
        "gender",
        "locale",
        "relationship_status"
    ]
    label_col = "gemini_score"
    
    def __init__(self):
        self.model = None
        self.encoders = {}
        self.scaler = None
        self.embedding_dim = 768
        self.logger = logging.getLogger(__name__)
        
        self.use_gpu = self._detect_gpu_availability()
        if self.use_gpu:
            self.logger.info("GPU detected - will use GPU acceleration")
        else:
            self.logger.info("GPU not available - using CPU fallback")

    def _detect_gpu_availability(self) -> bool:
        """Detect if GPU is available for XGBoost training"""
        try:
            test_data = xgb.DMatrix(np.random.rand(10, 5), label=np.random.rand(10))
            test_params = {
                "objective": "reg:squarederror",
                "tree_method": "hist",
                "device": "cuda",
                "verbosity": 0
            }
            xgb.train(test_params, test_data, num_boost_round=1, verbose_eval=False)
            return True
        except (xgb.core.XGBoostError, Exception) as e:
            self.logger.debug(f"GPU detection failed: {e}")
            return False

    def _validate_embedding(self, emb):
        try:
            arr = np.array(emb, dtype=np.float32)  # Use float32 for memory efficiency
            if arr.ndim == 1:
                return arr
            elif arr.ndim == 2:
                return arr.flatten()
            else:
                raise ValueError
        except Exception:
            self.logger.warning("Invalid embedding found, replaced with zeros.")
            return np.zeros(self.embedding_dim, dtype=np.float32)

    def _prepare_features(self, df: pd.DataFrame) -> np.ndarray:
        for feature in self.required_features:
            if feature not in df.columns:
                raise ValueError(f"Missing required feature: {feature}")

        X_emb = np.vstack([self._validate_embedding(emb) for emb in df["embedding"].values])
        
        if self.scaler is None:
            self.scaler = StandardScaler()
            X_emb = self.scaler.fit_transform(X_emb)
        else:
            X_emb = self.scaler.transform(X_emb)
        
        cat_features = ["gender", "locale", "relationship_status"]
        X_cat = []
        for col in cat_features:
            filled = df[col].fillna("(null)")
            if col not in self.encoders:
                self.encoders[col] = LabelEncoder()
                X_cat.append(self.encoders[col].fit_transform(filled))
            else:
                unseen_mask = ~filled.isin(self.encoders[col].classes_)
                if unseen_mask.any():
                    self.logger.warning(f"Found unseen labels in column '{col}': {filled[unseen_mask].unique()}")
                    default_label = "(null)" if "(null)" in self.encoders[col].classes_ else self.encoders[col].classes_[0]
                    filled = filled.where(~unseen_mask, default_label)
                X_cat.append(self.encoders[col].transform(filled))
        X_cat = np.vstack(X_cat).T
        
        def get_age(bday):
            try:
                parts = bday.split("/")
                if len(parts) == 3:
                    from datetime import date
                    current_year = date.today().year
                    return current_year - int(parts[2])
            except Exception as _:
                return np.nan
        X_age = np.array(df["birthday"].fillna("").apply(get_age).fillna(-1).values).reshape(-1, 1)
        X = np.hstack([X_emb, X_cat.astype(np.float32), X_age.astype(np.float32)])  # Ensure float32
        return X.astype(np.float32)  # Convert entire feature matrix to float32 for GPU memory efficiency

    def load_data(self, df: pd.DataFrame, label_col="gemini_score"):
        
        df = df.copy()
        df['score_bin'] = pd.qcut(df[label_col], q=5, duplicates='drop')

        try:
            train_df, test_df = train_test_split(
                df, 
                test_size=0.2, 
                random_state=42,
                stratify=df['score_bin']
            )
            self.logger.info("Used stratified sampling for train/test split")
        except ValueError as e:
            self.logger.warning(f"Stratified sampling failed: {e}. Using random split.")
            train_df, test_df = train_test_split(df, test_size=0.2, random_state=42)
        
        train_df = train_df.drop('score_bin', axis=1)
        test_df = test_df.drop('score_bin', axis=1)
        
        self.X_train = self._prepare_features(train_df)
        self.y_train = train_df[label_col].values.astype(np.float32)  # float32 for consistency
        self.X_test = self._prepare_features(test_df)
        self.y_test = test_df[label_col].values.astype(np.float32)

    def _get_base_params(self) -> dict:
        """Get base parameters with automatic GPU/CPU selection"""
        base = {
            "objective": "reg:squarederror",
            "eval_metric": "rmse",
            "tree_method": "hist",  # Modern parameter for both CPU and GPU
            "seed": 42,
        }
        if self.use_gpu:
            base["device"] = "cuda"
        else:
            base["device"] = "cpu"
        return base

    def auto_tune(self):
        if not hasattr(self, "X_train"):
            raise ValueError("Data not loaded. Call load_data first.")

        base_params = self._get_base_params()
        
        sample_size = min(12000, len(self.X_train))
        X_sample = self.X_train[:sample_size]
        y_sample = self.y_train[:sample_size]
        
        def objective(trial):
            """Optuna objective function for hyperparameter optimization"""
            try:
                params = base_params.copy()
                params.update({
                    "booster": trial.suggest_categorical("booster", ["gbtree", "dart"]),
                    "grow_policy": trial.suggest_categorical("grow_policy", ["depthwise", "lossguide"]),
                    "verbosity": 0,
                    "eta": trial.suggest_float("eta", 0.03, 0.2, log=True),
                    "max_depth": trial.suggest_int("max_depth", 4, 9),
                    "min_child_weight": trial.suggest_int("min_child_weight", 1, 6),
                    "subsample": trial.suggest_float("subsample", 0.7, 1.0),
                    "colsample_bytree": trial.suggest_float("colsample_bytree", 0.7, 1.0),
                    "gamma": trial.suggest_float("gamma", 0, 0.3),
                    "reg_alpha": trial.suggest_float("reg_alpha", 0, 1.0),
                    "reg_lambda": trial.suggest_float("reg_lambda", 0.8, 2.0),
                    "lr_decay": trial.suggest_float("lr_decay", 0.8, 1)
                })
                
                n_estimators = trial.suggest_int("n_estimators", 100, 800)
                
                dtrain = xgb.QuantileDMatrix(X_sample, label=y_sample)  # Optimized for hist binning on GPU
                
                pruning_callback = XGBoostPruningCallback(trial, "test-rmse-mean")
                lrdecay_callback = xgb.callback.LearningRateScheduler(
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
                    callbacks=[pruning_callback, lrdecay_callback],
                    verbose_eval=False,
                )
                
                best_rmse = cv_results["test-rmse-mean"].min()
                return best_rmse
                  
            except Exception as e:
                self.logger.warning(f"Trial failed: {e}")
                return float('inf')
        
        try:
            study = optuna.create_study(
                direction="minimize",
                study_name=f"xgboost_tuning_{datetime.now().strftime('%Y%m%d_%H%M%S')}",
                sampler=optuna.samplers.TPESampler(
                    seed=42,
                    n_startup_trials=10,
                    n_ei_candidates=24,
                    multivariate=True,
                    group=True,
                ),
                pruner=optuna.pruners.MedianPruner(
                    n_startup_trials=5,
                    n_warmup_steps=10,
                    interval_steps=5
                ),
            )
            
            study.optimize(
                objective, 
                n_trials=100,
                timeout=7200,
                show_progress_bar=False
            )
            
            best_params = study.best_params
            self.logger.info(f"Optuna optimization completed. Best RMSE: {study.best_value:.4f}")
            self.logger.info(f"Best parameters: {best_params}")
            
            return {
                "eta": best_params["eta"],
                "max_depth": best_params["max_depth"],
                "min_child_weight": best_params["min_child_weight"],
                "subsample": best_params["subsample"],
                "colsample_bytree": best_params["colsample_bytree"],
                "gamma": best_params["gamma"],
                "reg_alpha": best_params["reg_alpha"],
                "reg_lambda": best_params["reg_lambda"],
                "n_estimators": best_params["n_estimators"],
            }
              
        except Exception as e:
            self.logger.warning(f"Optuna optimization failed with {'GPU' if self.use_gpu else 'CPU'}: {e}")
            if self.use_gpu:
                self.logger.info("Falling back to CPU for hyperparameter optimization")
                self.use_gpu = False
                return self.auto_tune()
            else:
                self.logger.error("CPU optimization also failed, using default parameters")
                return {
                    "eta": 0.1,
                    "max_depth": 6,
                    "subsample": 0.8,
                    "colsample_bytree": 0.8,
                    "n_estimators": 500,
                }

    def train(self, auto_tune: bool = False):
        dtrain = xgb.QuantileDMatrix(self.X_train, label=self.y_train)  # Optimized for GPU hist
        dval = xgb.DMatrix(self.X_test, label=self.y_test)

        params = self._get_base_params()
        params.update({
            "eta": 0.1,
            "max_depth": 6,
            "subsample": 0.8,
            "colsample_bytree": 0.8,
        })
        
        num_boost_round = 500
        
        if auto_tune:
            try:
                best_params = self.auto_tune()
                if "n_estimators" in best_params:
                    num_boost_round = best_params.pop("n_estimators")
                params.update(best_params)
                self.logger.info(f"Using Optuna-tuned parameters: {best_params}")
            except Exception as e:
                self.logger.warning(f"Optuna tuning failed, using default parameters: {e}")

        try:
            self.logger.info(f"Training model with {'GPU' if self.use_gpu else 'CPU'}")
            self.model = xgb.train(
                params,
                dtrain,
                num_boost_round=num_boost_round,
                evals=[(dtrain, "train"), (dval, "eval")],
                early_stopping_rounds=20,
                verbose_eval=False
            )
            self.logger.info("Model training completed successfully")
              
        except Exception as e:
            if self.use_gpu:
                self.logger.warning(f"GPU training failed: {e}")
                self.logger.info("Falling back to CPU training")
                self.use_gpu = False
                params.update(self._get_base_params())
                
                self.model = xgb.train(
                    params,
                    dtrain,
                    num_boost_round=num_boost_round,
                    evals=[(dtrain, "train"), (dval, "eval")],
                    early_stopping_rounds=20,
                    verbose_eval=False
                )
                self.logger.info("Model training completed successfully with CPU fallback")
            else:
                self.logger.error(f"CPU training also failed: {e}")
                raise e
        
    def test(self):
        if self.model is None:
            raise ValueError("No model to test")

        if not hasattr(self, 'X_test') or not hasattr(self, 'y_test'):
            raise ValueError("Test data not loaded")

        dtest = xgb.DMatrix(self.X_test)
        y_pred = self.model.predict(dtest)
        rmse = root_mean_squared_error(self.y_test, y_pred)

        self.test_results = {
            "rmse": rmse,
            "r2": r2_score(self.y_test, y_pred),
            "mae": np.mean(np.abs(self.y_test - y_pred)),
        }
        # Convert numpy types to Python native types for JSON serialization
        self.test_results = self._convert_numpy_types(self.test_results)
        return self.test_results
      
    def load_model(self, model_name: str):
        model_dir = os.path.join(self.model_path, model_name)
        
        self.model = xgb.Booster()
        self.model.load_model(os.path.join(model_dir, "model.json"))
        
        import pickle
        encoders_path = os.path.join(model_dir, "encoders.pkl")
        if os.path.exists(encoders_path):
            with open(encoders_path, "rb") as f:
                self.encoders = pickle.load(f)
            self.logger.info(f"Encoders loaded from {encoders_path}")
        else:
            self.logger.warning(f"Encoders file not found at {encoders_path}. Model loaded without encoders.")
            self.encoders = {}
        
        scaler_path = os.path.join(model_dir, "scalers.pkl")
        if os.path.exists(scaler_path):
            with open(scaler_path, "rb") as f:
                self.scaler = pickle.load(f)
            self.logger.info(f"Scaler loaded from {scaler_path}")
        else:
            self.logger.warning(f"Scaler file not found at {scaler_path}. Model loaded without scaler.")
            self.scaler = None
            
        self.logger.info(f"Model loaded from {model_dir}")
      
    def predict(self, df: pd.DataFrame) -> list:
        if self.model is None:
            raise ValueError("No model found")
        X = self._prepare_features(df)
        dmatrix = xgb.DMatrix(X)
        predictions = self.model.predict(dmatrix)  # Will use GPU if model was trained on GPU and device=cuda
        # Convert numpy array to Python list for JSON serialization
        return self._convert_numpy_types(predictions)
    
    def _convert_numpy_types(self, obj):
        """Recursively convert numpy types to Python native types for JSON serialization"""
        if isinstance(obj, dict):
            return {key: self._convert_numpy_types(value) for key, value in obj.items()}
        elif isinstance(obj, list):
            return [self._convert_numpy_types(item) for item in obj]
        elif isinstance(obj, np.integer):
            return int(obj)
        elif isinstance(obj, np.floating):
            return float(obj)
        elif isinstance(obj, np.ndarray):
            return obj.tolist()
        elif isinstance(obj, (np.bool_, bool)):
            return bool(obj)
        else:
            return obj
      
    def save_model(self, model_name: str):
        if self.model is None:
            raise ValueError("No model to save")
        model_dir = os.path.join(self.model_path, model_name)
        os.makedirs(model_dir, exist_ok=True)
        
        self.model.save_model(os.path.join(model_dir, "model.json"))
        
        import pickle
        with open(os.path.join(model_dir, "encoders.pkl"), "wb") as f:
            pickle.dump(self.encoders, f)
        
        if self.scaler is not None:
            with open(os.path.join(model_dir, "scalers.pkl"), "wb") as f:
                pickle.dump(self.scaler, f)
        
        # Prepare metadata with numpy type conversion
        if not hasattr(self, 'test_results'):
            self.test_results = {}
        self.test_results["saved_at"] = datetime.now().isoformat()
        self.test_results["is_gpu"] = self.use_gpu
        
        # Convert numpy types to Python native types for JSON serialization
        metadata = self._convert_numpy_types(self.test_results)
        
        with open(os.path.join(model_dir, "metadata.json"), "w") as f:
            json.dump(metadata, f, indent=2)
            
        self.logger.info(f"Model saved to {model_dir}")