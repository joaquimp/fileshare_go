# Build stage
FROM golang:1.21-alpine AS builder

# Instalar certificados CA e git
RUN apk --no-cache add ca-certificates git

# Definir diretório de trabalho
WORKDIR /app

# Copiar go mod e sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copiar código fonte
COPY . .

# Build da aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Production stage
FROM alpine:latest

# Instalar certificados CA
RUN apk --no-cache add ca-certificates tzdata

# Criar usuário não-root
RUN addgroup -g 1001 appgroup && \
    adduser -D -s /bin/sh -u 1001 -G appgroup appuser

# Definir diretório de trabalho
WORKDIR /app

# Copiar binário da aplicação do builder stage
COPY --from=builder /app/main .

# Criar diretório uploads e dar permissões
RUN mkdir -p ./uploads && \
    chown -R appuser:appgroup /app

# Mudar para usuário não-root
USER appuser

# Expor porta
EXPOSE 8080

# Comando para executar a aplicação
CMD ["./main"]
