┌─────────────┐
│   Usuario   │
└──────┬──────┘
│  (auth + permisos)
▼
┌─────────────┐
│   Team      │  ← equipo / proyecto
└──────┬──────┘
│
▼
┌─────────────┐
│ Namespace   │  ← aislamiento real
└──────┬──────┘
│
▼
┌─────────────┐
│   Apps      │  ← deployments / services
└─────────────┘

🔐 Principio de seguridad clave
- Los usuarios NO hablan directamente con Kubernetes

✅ Tus usuarios interactúan solo con tu API
✅ Kubernetes solo confía en:
- kubeconfig admin (de tu API)
- ServiceAccounts controlados por ti

Esto elimina:
- RBAC complejo con usuarios reales
- gestión de certificados
- ataques laterales