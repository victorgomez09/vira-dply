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
          └─────────────┬─────────────────┘
                        │
                        ▼
          ┌───────────────────────────────┐
          │ Coroutine (Dispatchers.IO)    │
          │ provisionCluster(env)         │
          └─────────────┬─────────────────┘
                        │
      ┌─────────────────┴─────────────────────┐
      │                                         │
      ▼                                         ▼
┌──────────────┐                         ┌──────────────┐
│ k3d CLI      │                         │ Kubeconfig   │
│ - cluster    │                         │ SecretStore  │
│   create     │                         │ - guarda    │
│ - kubeconfig │                         │   cifrado    │
└─────┬────────┘                         └─────┬────────┘
│                                        │
▼                                        ▼
┌──────────────┐                         ┌──────────────┐
│ client-java  │                         │ Database     │
│ CoreV1Api    │                         │ - Environment│
│ valida nodos │                         │   status     │
└─────┬────────┘                         └──────────────┘
│
▼
┌──────────────┐
│ Update BD    │
│ - status=READY│
│ - kubeconfigRef│
└──────────────┘
│
▼
┌──────────────┐
│ GET /environments/{id} │
│ -> Environment.status   │
└─────────────────────────┘
