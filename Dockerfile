# Build stage
FROM golang:1.21-alpine3.19 AS builder

# Atualizar pacotes e instalar certificados CA e git
RUN apk update && apk upgrade && apk --no-cache add ca-certificates git

# Definir diretório de trabalho
WORKDIR /app

# Copiar go mod e sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copiar código fonte
COPY . .

# Build da aplicação principal
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Build do healthcheck
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o healthcheck healthcheck.go

# Production stage
FROM alpine:3.19

# Atualizar pacotes e instalar certificados CA
RUN apk update && apk upgrade && apk --no-cache add ca-certificates tzdata

# Criar usuário não-root
RUN addgroup -g 1001 appgroup && \
    adduser -D -s /bin/sh -u 1001 -G appgroup appuser

# Definir diretório de trabalho
WORKDIR /app

# Copiar binários da aplicação do builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/healthcheck .

# Criar diretório uploads e dar permissões
RUN mkdir -p ./uploads && \
    chown -R appuser:appgroup /app

# Mudar para usuário não-root
USER appuser

# Expor porta
EXPOSE 8080

# Adicionar health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD ./healthcheck

# Comando para executar a aplicação
CMD ["./main"]
