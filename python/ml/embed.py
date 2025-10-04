import logging

import numpy as np
import torch
from FlagEmbedding import BGEM3FlagModel


class BGEM3EmbedModel:
  MODEL_USED = "BAAI/bge-m3"

  def __init__(self):
    self.logger = logging.getLogger(__name__)
    if torch.cuda.is_available():
      device = "cuda:0"
      gpu_name = torch.cuda.get_device_name(0)
      gpu_memory = torch.cuda.get_device_properties(0).total_memory / 1024**3
      self.logger.info("Using GPU: %s", gpu_name)
      self.logger.info("GPU Memory: %.2f GB", gpu_memory)
    else:
      device = "cpu"
      self.logger.warning("GPU not available, falling back to CPU")
    self.model = BGEM3FlagModel(
      self.MODEL_USED,
      use_fp16=True,
      devices=device,
    )
    self.logger.info("BGE-M3 model loaded on %s", device)

  def embed(self, texts: str | list[str]) -> list[float] | list[list[float]]:
    """
    Embed text(s) using BGE-M3 model.

    Args:
        texts: Single text string or list of texts

    Returns:
        Single embedding vector or list of embedding vectors

    """
    if isinstance(texts, str):
      texts = [texts]
      return_single = True
    else:
      return_single = False

    embeddings = self.model.encode(
        texts,
        batch_size=256,
        max_length=1024
    )

    dense_vecs = embeddings["dense_vecs"]

    if isinstance(dense_vecs, np.ndarray):
        result: list[list[float]] = dense_vecs.tolist()
    else:
        result: list[list[float]] = [[float(x) for x in vec] for vec in dense_vecs]

    return result[0] if return_single else result
