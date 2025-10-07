# E-Commerce Microservices Backend System (Go + CockroachDB + Kubernetes)

## ● Overview
This project implements a **microservices-based e-commerce backend** built using **Go**, orchestrated via **Kubernetes**, and backed by **CockroachDB**.  
It includes three main services: **Authentication**, **Product Management**, and **Order Processing**.  
All services expose **REST APIs** and are independently deployable using **Docker** and **Kubernetes**.

---

## ● System Architecture

### - Components

| Service | Port | Description |
|----------|------|-------------|
| Auth Service | 8082 | User registration, login, and JWT token management |
| Product Service | 8084 | CRUD operations on products |
| Order Service | 8083 | Order creation and management |
| CockroachDB | 26257 (SQL), 8080 (Admin UI) | Distributed SQL database for all services |

**Data Flow**
1. Users register or login using the Auth Service → receive JWT token.
2. Product and Order Services require this JWT for secure operations.
3. All services share a single CockroachDB database with separate tables.
4. Kubernetes handles deployment, scaling, and zero-downtime rollouts.

---

## ● Technologies Used
- **Go (Golang)** – microservices implementation
- **CockroachDB** – distributed SQL database
- **Docker** – containerization
- **Kubernetes** – orchestration and deployment

---

## ● Project Structure

```
E-commerce-microservice/
├── services/
│   ├── auth/
│   ├── order/
│   ├── product/
│
├── common/                # Shared utilities and modules
│   ├── auth/              # Common JWT authentication logic, token/hash generation logic 
│   ├── http/              # http middleware logic 
│   ├── migrate/           # Common CockroachDB migration logic
│   └── response/          # Standardized response helpers
│
├── misc/                  # Testing and API documentation utilities
│   ├── test.sh            # E2E system test script
│   └── routes.json        # Postman export (all API endpoints)
│
├── k8s/                   # Kubernetes manifests
│   ├── 00-namespace.yaml
│   ├── 10-cockroachdb.yaml
│   ├── 20-config-auth.yaml
│   ├── 21-config-order.yaml
│   ├── 22-config-product.yaml
│   ├── 30-secrets.yaml
│   ├── 40-auth.yaml
│   ├── 41-order.yaml
│   ├── 42-product.yaml
│   └── 80-ingress.yaml
│
└── README.md
```

---

## ● Common Folder

The **`common/`** folder centralizes reusable logic and shared packages between all services, improving modularity and reducing duplication.

| Subfolder     | Purpose                                                                                |
|---------------|----------------------------------------------------------------------------------------|
| **auth/**     | Contains JWT authentication middleware and logic for token/hash generation/decryption. |
| **migrate/**  | Handles CockroachDB connections, schema initialization and migration.                  |
| **http/**     | Handle common middleware logic to authenticate and handle requests.                    |
| **response/** | Provides standardized response and error formatting functions.                         |

These shared components are imported by each microservice for consistent behavior.

---

## ● Misc Folder

The **`misc/`** folder contains helper files for testing, verification, and API documentation.

- **`test.sh`** → Automated end-to-end testing script covering complete system behavior.
- **`routes.json`** → Postman collection containing all API endpoints (Auth, Product, Order).
---

## ● Docker Setup

Each microservice contains its own **Dockerfile**.

### Build Images
```bash
   docker build -t ecommerce-auth:latest ./services/auth
docker build -t ecommerce-order:latest ./services/order
docker build -t ecommerce-product:latest ./services/product
```

---

## ● Kubernetes Deployment

### - Apply Namespace
```bash
   kubectl apply -f k8s/00-namespace.yaml
```

### - Deploy CockroachDB
```bash
   kubectl apply -f k8s/10-cockroachdb.yaml
kubectl -n ecommerce-ms get pods   # wait for DB ready
```

### - Apply Configs and Secrets
```bash
   kubectl apply -f k8s/20-config-auth.yaml -f k8s/21-config-order.yaml -f k8s/22-config-product.yaml
kubectl apply -f k8s/30-secrets.yaml
```

### - Deploy Services
```bash
   kubectl apply -f k8s/40-auth.yaml -f k8s/41-order.yaml -f k8s/42-product.yaml
```

### - Enable Ingress and Autoscaling
```bash
   kubectl apply -f k8s/80-ingress.yaml
kubectl apply -f k8s/90-hpa-example.yaml
```

---

## ● Zero-Downtime Deployment
- Each service has **2 replicas**.
- Rolling updates are configured:
  ```yaml
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 0
      maxSurge: 1
  ```
- Liveness & Readiness probes ensure pods receive traffic only when ready.

---

## ● Security Notes
- JWT tokens for all protected routes.
- Database not exposed externally.
- Secrets handled via Kubernetes `app-secrets`.

---

## ● Future Enhancements
- Add Redis for caching and sessions.
- Integrate CI/CD pipeline for automated builds.
- Enable HTTPS via NGINX Ingress annotations.

---

## ● Summary
✅ Modular Go microservices (Auth, Product, Order)  
✅ Shared reusable logic via `common/` package  
✅ REST APIs secured with JWT  
✅ CockroachDB integration  
✅ Docker + Kubernetes orchestration  
✅ Zero-downtime deployment with rolling updates  
✅ Fully tested via `misc/test.sh` and documented via Postman collection

---

## ● Author

Made by Harshvardhan Rajpurohit – Backend Developer • Go / Node.js • GCP • Kubernetes • Clean Architecture
