"""Financial Analysis model for storing Gemini analysis results"""

from datetime import datetime
from typing import TYPE_CHECKING, Optional
from sqlalchemy import ForeignKey, Integer, String, Text, DateTime, Float, JSON
from sqlalchemy.orm import mapped_column, Mapped, relationship
from pydantic import BaseModel, Field

from .base import Base

if TYPE_CHECKING:
  from .profile import UserProfile
  from .prompt import Prompt


class FinancialAnalysis(Base):
  """Financial Analysis model for storing Gemini analysis results"""

  __tablename__ = "financial_analysis"

  id: Mapped[int] = mapped_column(Integer, primary_key=True)

  financial_status: Mapped[str] = mapped_column(
    String, nullable=False, comment="Overall financial status: 'low', 'medium', 'high'"
  )
  confidence_score: Mapped[float] = mapped_column(
    Float, nullable=False, comment="Confidence score between 0.0 and 1.0"
  )
  analysis_summary: Mapped[str] = mapped_column(
    Text, nullable=False, comment="Summary of the analysis"
  )
  indicators: Mapped[Optional[dict]] = mapped_column(
    JSON, nullable=True, comment="JSON object containing specific indicators found"
  )

  gemini_model_used: Mapped[str] = mapped_column(String, nullable=False)
  prompt_tokens_used: Mapped[Optional[int]] = mapped_column(Integer, nullable=True)
  prompt_used_id: Mapped[int] = mapped_column(ForeignKey("prompt.id"), nullable=False)
  prompt_used: Mapped["Prompt"] = relationship()
  completion_tokens_used: Mapped[Optional[int]] = mapped_column(Integer, nullable=True)
  total_tokens_used: Mapped[Optional[int]] = mapped_column(Integer, nullable=True)

  created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.now)
  updated_at: Mapped[datetime] = mapped_column(
    DateTime, default=datetime.now, onupdate=datetime.now
  )

  user_profile_id: Mapped[int] = mapped_column(ForeignKey("user_profile.id"), nullable=False)
  user_profile: Mapped["UserProfile"] = relationship(back_populates="financial_analyses")

  def to_schema(self) -> "FinancialAnalysisSchema":
    """Convert to schema"""
    return FinancialAnalysisSchema.model_validate(self)

  def to_json(self) -> dict:
    """Convert to JSON serializable dict"""
    return FinancialAnalysisSchema.model_validate(self).model_dump()


class FinancialAnalysisSchema(BaseModel):
  """Schema for FinancialAnalysis model"""

  id: int
  financial_status: str
  confidence_score: float
  analysis_summary: str
  indicators: Optional[dict] = None
  gemini_model_used: str
  prompt_tokens_used: Optional[int] = None
  completion_tokens_used: Optional[int] = None
  total_tokens_used: Optional[int] = None
  created_at: datetime
  updated_at: datetime
  user_profile_id: int

  model_config = {"from_attributes": True}


class FinancialAnalysisCreateDTO(BaseModel):
  """Data Transfer Object for creating a new financial analysis"""

  financial_status: str = Field(..., pattern=r"^(low|medium|high)$")
  confidence_score: float = Field(..., ge=0.0, le=1.0)
  analysis_summary: str
  indicators: Optional[dict] = None
  gemini_model_used: str
  prompt_tokens_used: Optional[int] = None
  completion_tokens_used: Optional[int] = None
  total_tokens_used: Optional[int] = None


class BatchAnalysisRequest(BaseModel):
  """Request for batch analysis of multiple profiles"""

  profile_ids: list[int] = Field(..., min_length=1, max_length=10)
  force_reanalysis: bool = Field(
    default=False, description="Force re-analysis even if recent analysis exists"
  )


class BatchAnalysisResponse(BaseModel):
  """Response for batch analysis"""

  success: bool
  results: list[FinancialAnalysisSchema]
  errors: list[dict] = Field(default_factory=list)
  total_tokens_used: int = 0
