# Charts API Documentation

## Endpoints đã được thêm vào branch feature/charts-api:

### 1. Dashboard Statistics
```
GET /data/chart/dashboard
```
**Response:**
```json
{
  "data": {
    "total_groups": 150,
    "total_comments": 12450,
    "total_posts": 3240,
    "total_profiles": 8500,
    "embedded_count": 4800,
    "scanned_profiles": 7200,
    "scored_profiles": 4200,
    "analyzed_profiles": 5100,
    "total_accounts": 25,
    "active_accounts": 18,
    "blocked_accounts": 7
  }
}
```

### 2. Time Series Data
```
GET /data/chart/timeseries
```
**Response:**
```json
{
  "data": [
    {
      "date": "2024-10-01",
      "count": 1200,
      "data_type": "profiles"
    },
    {
      "date": "2024-10-01", 
      "count": 450,
      "data_type": "posts"
    },
    {
      "date": "2024-10-01",
      "count": 1800,
      "data_type": "comments"
    }
  ]
}
```

### 3. Score Distribution
```
GET /data/chart/scores
```
**Response:**
```json
{
  "data": [
    {
      "score_range": "0.0-0.2",
      "count": 850,
      "percentage": 20.2
    },
    {
      "score_range": "0.2-0.4",
      "count": 1200,
      "percentage": 28.6
    }
  ]
}
```

## SQL Queries được thêm:

1. **GetDashboardStats** - Thống kê tổng quan toàn bộ hệ thống
2. **GetTimeSeriesData** - Dữ liệu theo thời gian (6 tháng gần nhất)
3. **GetScoreDistribution** - Phân bố điểm số Gemini

## Files đã thay đổi:

- `infras/sql/query.sql` - Thêm 3 queries mới
- `server/modules/routes/data.go` - Thêm 3 endpoints
- `server/modules/routes/services/data/data.go` - Thêm 3 service methods
- `db/query.sql.go` - Auto-generated từ sqlc

## Cách test:

1. Start server: `go run main.go` hoặc `./dev.ps1`
2. Test endpoints:
   ```bash
   curl http://localhost:8000/data/chart/dashboard
   curl http://localhost:8000/data/chart/timeseries
   curl http://localhost:8000/data/chart/scores
   ```

## Tích hợp Frontend:

Trong asfpc-ui, có thể sử dụng RTK Query để gọi các endpoints này và tạo charts với Recharts.