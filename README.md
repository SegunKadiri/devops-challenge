# DevOps Challenge — Complete Kubernetes + Golang Monitoring Solution

This repository contains a fully working implementation of the DevOps challenge using **Kind**, **Kubernetes**, **Helm**, **Nginx**, **MariaDB**, **NetworkPolicies**, **CronJobs**, and a custom **Golang pod monitoring application**.

Everything is deployable locally — **no cloud accounts needed**.

---

# Architecture
+---------------------------+
|         Browser           |
+-------------+-------------+
              |
              v
+-------------+------------+         +---------------------------+
|      Web Deployment      |  --->   |      NetworkPolicy        |
|  Nginx with InitContainer|         |   Allow only Web → DB     |
|  Shows Pod IP + Serving  |         +---------------------------+
|  Host = Host-xxxxx       |
+-------------+------------+
              |
              v
+-------------+------------+
|  MySQL StatefulSet       |
|  Persistent Volume (PVC) |
|  Node Affinity for DB    |
+-------------+------------+
              |
              v
+-------------+------------+
| DB Backup CronJob        |
| Disaster Recovery (DR)   |
+--------------------------+

+--------------------------------------------------------------+
|                 Go Pod Monitor Deployment                    |
| Watches ALL K8s pod events: ADDED, UPDATED, DELETED          |
| Prints structured logs with Pod, Namespace, Phase, IP        |
+--------------------------------------------------------------+


# 📌 Features Implemented (Checklist)

### ✅ 1. Kubernetes cluster (Kind, multi-node)  
### ✅ 2. MariaDB DB cluster with persistent volumes (StatefulSet)  
### ✅ 3. Web server (Nginx) with:
- Multiple replicas  
- Custom Nginx config  
- Init container modifying HTML  
- Webpage displays Pod IP  
- Webpage displays `serving-host = Host-xxxxx` where `xxxxx = last 5 chars of pod name`

### ✅ 4. NetworkPolicy  
Only web pods are allowed to communicate with the DB on port 3306.

### ✅ 5. Disaster Recovery  
A CronJob runs periodic database backups using `mysqldump` into a PVC.

### ✅ 6. Flexible pod multi-network support  
Pods may optionally attach to a second network using Multus annotations.

### ✅ 7. Node scheduling rules  
Database replicas can be scheduled on specific nodes using `nodeAffinity`.

### ✅ 8. Golang Monitoring Application  
A Go app (running inside Kubernetes) monitors **all pod events**:
- ADDED  
- UPDATED (phase changes)  
- DELETED  

### ✅ 9. Helm Chart Deploys Everything  
All components are shipped via a single Helm release.

---

# 📂 Repository Structure

```
.
├── charts/
│   └── devops-demo/        # Main Helm chart (DB, Web, Monitor, Policies, CronJob)
│
├── go-monitor/             # Golang pod monitoring app (client-go)
│   ├── main.go
│   ├── go.mod
│   └── Dockerfile
│
├── scripts/                # Optional helper scripts
│   ├── build-images.sh
│   ├── deploy.sh
│   ├── test-all.sh
│   ├── destroy-cluster.sh
│   └── ...
│
├── kind-config.yaml        # Multi-node Kind cluster definition
└── README.md               # This file
```

# ⚙️ Prerequisites

- **Docker** ≥ 24.x  
- **Kind** ≥ 1.29.x  
- **kubectl** ≥ 1.34.x  
- **Helm** ≥ 3.x  
- **Go** ≥ 1.21 (only required to build monitor locally)

> Optional: Multus CNI if testing multi-network support

---

# 🚀 Deployment Instructions

You can deploy either:

- **Using the scripts** (recommended for convenience)  
or  
- **Manually using raw commands**

Both approaches are supported.

---

# OPTION 1 — Deploy Using Scripts (Recommended)

From the repo root:

### 1. Create cluster + build images + install via Helm
```bash
./scripts/deploy.sh
```

### 2. Run automated tests
```bash
./scripts/test-all.sh
```

### 3. Tear down cluster
```bash
./scripts/destroy-cluster.sh
```

These scripts are optional — you may modify or extend them later.

---

# OPTION 2 — Manual Deployment (Full Transparency)

### 1. Create Kind cluster
```bash
kind create cluster --name devops-challenge --config kind-config.yaml
```

### 2. Build Golang monitor locally
```bash
cd go-monitor
docker build -t devops-monitor:local .
```

### 3. Load image into Kind nodes
```bash
kind load docker-image devops-monitor:local --name devops-challenge
```

### 4. Install everything using Helm
From repo root:
```bash
helm upgrade --install devops-demo charts/devops-demo --wait --timeout 5m
```

---

# 🌐 Testing the Web Server

When Nginx deploys, access it with:

```
http://localhost:30080
```

The page shows:

- Pod IP  
- Serving Host value  
- Nginx served HTML modified by init container  

Example:

```
Pod IP: 10.244.2.14
serving-host: Host-5gxj
```

---

# 🔐 Database NetworkPolicy Test

From a pod NOT labeled `app=devops-demo-web`:

```bash
mysql -h devops-demo-mysql.default.svc.cluster.local -u root -p
```

Expected:

❌ Connection blocked by NetworkPolicy.

From Nginx web pod:

```bash
mysql -h devops-demo-mysql.default.svc.cluster.local -u root -p
```

Expected:

✅ Allowed.

---

# 💾 Database Disaster Recovery Test

The CronJob creates periodic dumps at:

```
/var/lib/mysql-backups/
```

To trigger manually:

```bash
kubectl create job --from=cronjob/devops-demo-mysql-backup manual-backup
```

---

# 🛰 Multi-Network Pod Test (Optional)

Pods can join a secondary network by adding:

```yaml
annotations:
  k8s.v1.cni.cncf.io/networks: macvlan-conf
```

---

# 🎯 Node Scheduling for DB Replicas

Nodes are labeled:

```bash
kubectl label node devops-challenge-worker db=node-1
kubectl label node devops-challenge-worker2 db=node-2
```

The StatefulSet includes:

```yaml
nodeAffinity:
  requiredDuringSchedulingIgnoredDuringExecution:
    nodeSelectorTerms:
      - matchExpressions:
          - key: db
            operator: In
            values:
              - node-1
```

Replica-specific scheduling is handled via StatefulSet ordinal rules.

---

# 🟦 Golang Pod Monitor Usage

Get logs:

```bash
kubectl logs -l app=devops-monitor --tail=200 -f
```

Example output:

```
2025-11-23T20:22:29Z - ADDED: default/devops-demo-web-75546bd9dd-9t4bm (IP:10.244.2.14)
2025-11-23T20:39:54Z - DELETED: default/devops-demo-web-75546bd9dd-hgnwg
2025-11-23T20:40:23Z - UPDATED: default/devops-demo-web-75546bd9dd-vtsd4 (phase=Failed)
```

---

# 🏭 Production-Level Considerations

While the current setup is designed for **local testing and development with Kind**, the following changes are recommended for a **production-ready Kubernetes deployment**:

### 1. Cluster Setup
- Use a **managed Kubernetes service** (EKS, GKE, AKS) or a production-grade **self-hosted cluster** instead of Kind.
- Enable **high availability** for control-plane nodes.
- Use **multiple worker nodes** with proper resource allocation.

### 2. Database (MariaDB/MySQL)
- Deploy **multi-AZ StatefulSet** with persistent volumes on production-grade storage (e.g., AWS EBS, GCP Persistent Disk).
- Enable **automated backups** to an external storage location (S3, GCS, etc.) for disaster recovery.
- Use **replication** for high availability.
- Configure **resource requests/limits** for DB pods.

### 3. Web Server (Nginx)
- Use **horizontal pod autoscaler (HPA)** to scale web replicas based on CPU/memory.
- Use **Ingress controller** for external access and TLS termination.
- Mount **ConfigMaps and Secrets** for Nginx configuration and sensitive data.
- Enable **readiness and liveness probes** for robust pod health monitoring.

### 4. Networking & Security
- Replace local **NetworkPolicy** testing with real **firewall and security group rules** in production.
- Use **RBAC** and **Service Accounts** with least privilege access.
- Consider **mutual TLS (mTLS)** or service mesh (Istio, Linkerd) for secure pod-to-pod communication.

### 5. Logging & Monitoring
- Forward logs to **centralized logging platform** (ELK, Loki, or Cloud provider logs).
- Monitor pod events using **Golang monitor**, but integrate with **Prometheus + Grafana** for production observability.
- Use **alerting** for pod failures, DB issues, or network anomalies.

### 6. Container Image Management
- Store images in a **private registry** (Docker Hub, GCR, ECR, or Harbor).
- Use **image tags for versioning**; avoid `:latest`.
- Implement **image scanning** for vulnerabilities.

### 7. Helm Charts & CI/CD
- Use **Helm values for environment-specific configuration** (dev, staging, prod).
- Automate deployment via **CI/CD pipelines** (GitHub Actions, GitLab CI, Jenkins).
- Include **unit/integration tests** for Helm charts and Go monitor application.

### 8. Optional Enhancements
- Enable **PodDisruptionBudgets (PDB)** for high availability.
- Use **PersistentVolumeClaims with storage classes** that support dynamic provisioning.
- Consider **multi-network setup** (via Multus) only if required by production networking topology.

---

> ⚠️ **Note:** These recommendations are intended to bridge the gap between a local Kind-based demo environment and a production-ready Kubernetes architecture.

