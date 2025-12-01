# Flujo entornos
          ┌───────────────────────────────┐
          │   Cliente (Frontend / API)   │
          │ POST /environments {name}    │
          └─────────────┬─────────────────┘
                        │
                        ▼
          ┌───────────────────────────────┐
          │ EnvironmentController         │
          │ - recibe request              │
          │ - llama createAndProvision()  │
          │ - devuelve 201 CREATED        │
          └─────────────┬─────────────────┘
                        │
                        ▼
          ┌───────────────────────────────┐
          │ EnvironmentService            │
          │ - guarda Environment en BD    │
          │   status = PROVISIONING       │
          │ - lanza coroutine async       │
          │ - registra logging / metrics │
          └─────────────┬─────────────────┘
                        │
                        ▼
          ┌───────────────────────────────┐
          │ Provisioning Coroutine        │
          │ - Dispatchers.IO              │
          │ - SupervisorJob (indep.)     │
          └─────────────┬─────────────────┘
                        │
          ┌─────────────┴─────────────────────────────┐
          │                                             │
          ▼                                             ▼
┌───────────────────────┐                       ┌───────────────────────┐
│ Retry / Backoff Loop  │                       │ Cancellation Token     │
│ - reintentos config.  │                       │ - permite cancelar    │
│ - registra logs       │                       │   provisión si necesario │
└─────────────┬─────────┘                       └─────────────┬─────────┘
│                                           │
▼                                           ▼
┌───────────────────────┐                 ┌─────────────────────────┐
│ k3d CLI               │                 │ Kubeconfig SecretStore  │
│ - cluster create      │                 │ - guarda cifrado        │
│ - kubeconfig get      │                 │ - delete / update       │
└─────────────┬─────────┘                 └─────────────┬──────────┘
│                                           │
▼                                           ▼
┌───────────────────────┐                 ┌─────────────────────────┐
│ client-java API       │                 │ Database                │
│ CoreV1Api validate    │                 │ - Environment status    │
│ nodos / readiness     │                 │ - kubeconfigRef         │
└─────────────┬─────────┘                 └─────────────┬──────────┘
│                                           │
▼                                           ▼
┌───────────────────────────────┐
│ Update Environment             │
│ - status = READY / FAILED      │
│ - updatedAt                    │
└─────────────┬─────────────────┘
│
▼
┌───────────────────────────────┐
│ GET /environments/{id}        │
│ -> Environment.status         │
└───────────────────────────────┘
