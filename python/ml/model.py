from datetime import datetime
import json
import logging
import os
import pandas as pd
from sklearn.calibration import LabelEncoder
from sklearn.metrics import root_mean_squared_error, r2_score
from sklearn.model_selection import train_test_split
import xgboost as xgb
import numpy as np

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
    self.logger = logging.getLogger(__name__)

  def _prepare_features(self, df: pd.DataFrame) -> np.ndarray:
    for feature in self.required_features:
      if feature not in df.columns:
        raise ValueError(f"Missing required feature: {feature}")
    
    X_emb = np.vstack(df["embedding"].values)  # Convert float[] sang numpy array
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
          return 2025 - int(parts[2])
        return np.nan
      except Exception as _:
          return np.nan
    X_age = df["birthday"].fillna("").apply(get_age).fillna(-1).values.reshape(-1, 1)
    X = np.hstack([X_emb, X_cat, X_age])
    return X

  def load_data(self, df: pd.DataFrame, label_col="gemini_score") -> pd.DataFrame:
    
    # Tạo stat key để phân bổ dữ liệu hợp lý hơn
    df['strat_key'] = df['locale'].fillna('unknown') + '_' + df['gender'].fillna('unknown')
    try:
      train_df, test_df = train_test_split(
        df, 
        test_size=0.2, 
        random_state=42,
        stratify=df['strat_key']
      )
      self.logger.info("Used stratified sampling for train/test split")
    except ValueError as e:
      self.logger.warning(f"Stratified sampling failed: {e}. Using random split.")
      train_df, test_df = train_test_split(df, test_size=0.2, random_state=42)
    
    train_df = train_df.drop('strat_key', axis=1)
    test_df = test_df.drop('strat_key', axis=1)
    
    self.X_train = self._prepare_features(train_df)
    self.y_train = train_df[label_col].values
    self.X_test = self._prepare_features(test_df)
    self.y_test = test_df[label_col].values
  
  def train(self):
    dtrain = xgb.DMatrix(self.X_train, label=self.y_train)

    # Tuning ở đây
    params = {
      "objective": "reg:squarederror",
      "eval_metric": "rmse",
      "eta": 0.1,
      "max_depth": 6,
      "subsample": 0.8,
      "colsample_bytree": 0.8,
      "seed": 42
    }

    self.model = xgb.train(
      params,
      dtrain,
      num_boost_round=200,
      evals=[(dtrain, "train")],
      early_stopping_rounds=20
    )
    
  def test(self) -> dict[str, str | float]:
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
    return self.test_results
  
  def load_model(self, model_name: str):
    model_dir = os.path.join(self.model_path, model_name)
    
    # Load the XGBoost model
    self.model = xgb.Booster()
    self.model.load_model(os.path.join(model_dir, "model.json"))
    
    # Load the encoders
    import pickle
    encoders_path = os.path.join(model_dir, "encoders.pkl")
    if os.path.exists(encoders_path):
      with open(encoders_path, "rb") as f:
        self.encoders = pickle.load(f)
      self.logger.info(f"Model and encoders loaded from {model_dir}")
    else:
      self.logger.warning(f"Encoders file not found at {encoders_path}. Model loaded without encoders.")
      self.encoders = {}
  
  def predict(self, df: pd.DataFrame) -> np.ndarray:
    if self.model is None:
      raise ValueError("No model found")
    X = self._prepare_features(df)
    dmatrix = xgb.DMatrix(X)
    predictions = self.model.predict(dmatrix)
    return predictions
  
  def save_model(self, model_name: str):
    if self.model is None:
      raise ValueError("No model to save")
    model_dir = os.path.join(self.model_path, model_name)
    os.makedirs(model_dir, exist_ok=True)
    self.model.save_model(os.path.join(model_dir, "model.json"))
    import pickle
    with open(os.path.join(model_dir, "encoders.pkl"), "wb") as f:
      pickle.dump(self.encoders, f)
    self.logger.info(f"Model saved to {model_dir}")
    if not hasattr(self, 'test_results'):
      self.test_results = {}
    self.test_results["saved_at"] = datetime.now().isoformat()
    with open(os.path.join(model_dir, "metadata.json"), "w") as f:
      json.dump(self.test_results, f)