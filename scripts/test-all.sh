#!/usr/bin/env bash
set -euo pipefail

echo "1) Check nodes"
kubectl get nodes -o wide

echo
echo "2) Check pods"
kubectl get pods -o wide

echo
echo "3) Check web endpoints (curl localhost:30080)"
curl -sS http://localhost:30080 || echo "curl failed"

echo
echo "4) Check mysql pods and PVCs"
kubectl get pods -l app=mysql-db -o wide
kubectl get pvc

echo
echo "5) Test network policy - from busybox pod (should fail)"
kubectl run -i --tty nettest --image=busybox:1.36 --restart=Never -- sh -c "nc -zv devops-demo-mysql 3306" || true
kubectl delete pod nettest || true

echo
echo "6) Show backup PVC contents (if any backups exist)"
BACKUP_POD=$(kubectl get pods -l job-name=devops-demo-mysql-backup -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
if [ -n "$BACKUP_POD" ]; then
  kubectl exec -it "$BACKUP_POD" -- ls /backup || true
else
  echo "No backup job pod currently running; check cronjob status: kubectl get cronjob"
fi

echo
echo "7) Tail monitor logs (5s)"
kubectl logs -l app=devops-monitor --tail=200 || echo "no monitor logs"
