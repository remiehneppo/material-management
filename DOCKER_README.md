# Material Management - Docker Setup

## Build và chạy với Docker Compose

### 1. Build và khởi chạy toàn bộ stack

```bash
docker-compose up --build
```

### 2. Chạy trong background

```bash
docker-compose up -d --build
```

### 3. Xem logs

```bash
# Xem logs của tất cả services
docker-compose logs -f

# Xem logs của app cụ thể
docker-compose logs -f material-management-app
```

### 4. Dừng services

```bash
docker-compose down
```

### 5. Dừng và xóa volumes (cẩn thận - sẽ mất data)

```bash
docker-compose down -v
```

## Cấu trúc Docker

### Services

- **material-management-app**: Ứng dụng Go chính
- **mongodb**: Database MongoDB với user admin/admin123
- **redis**: Cache Redis

### Volumes được mount

- `./logs:/root/logs` - Thư mục logs của ứng dụng
- `./uploads:/root/uploads` - Thư mục uploads
- `./config.yaml:/root/config.yaml` - File cấu hình

### Environment Variables

Tất cả các biến trong `.env` sẽ được truyền vào container:

- PORT
- MONGODB_URI
- MONGODB_DATABASE
- OPENAI_API_KEY
- ENVIRONMENT
- JWT_SECRET
- JWT_ISSUER
- JWT_EXPIRE
- REDIS_URL
- REDIS_USERNAME
- REDIS_PASSWORD

### Ports

- App: 8088 (có thể thay đổi qua PORT trong .env)
- MongoDB: 27017
- Redis: 6379

## Lưu ý

1. Đảm bảo file `.env` có đầy đủ các biến cần thiết
2. Thư mục `logs/` và `uploads/` sẽ được tạo tự động nếu chưa có
3. MongoDB sẽ tự động tạo database với thông tin trong MONGODB_DATABASE
4. Redis chạy mà không cần password (có thể uncomment dòng trong docker-compose.yml để thêm password)

## Build chỉ image của app

```bash
docker build -t material-management:latest .
```

## Chạy chỉ app (cần MongoDB và Redis đã chạy)

```bash
docker run --env-file .env -p 8088:8088 material-management:latest
```
