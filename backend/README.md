## 🔁 Flujo Completo de la Arquitectura CQRS/ES

El sistema opera en dos flujos principales, que están desacoplados por el **Message Broker (Kafka)**: el Flujo de Comando (Escritura) y el Flujo de Consulta (Lectura).

---

## 1. ✍️ Flujo de Comando (Escritura)

Este es el lado **transaccional** del sistema, centrado en el Agregado (`Product`) y la persistencia de eventos en **PostgreSQL/GORM**.

### A. Recepción y Manejo
| Paso | Componente | Acción Clave | Tecnología |
| :--- | :--- | :--- | :--- |
| **1.** | **API REST (Echo Handler)** | Recibe el `POST /products`. | Echo |
| **2.** | **Command Handler** (`CreateProductHandler`) | Orquesta la acción y llama al Dominio. | Aplicación (Go) |
| **3.** | **Agregado (Product)** | **Genera Evento:** Crea el agregado y produce el evento `ProductCreated`. | Dominio (DDD) |

### B. Persistencia y Publicación
| Paso | Componente | Acción Clave | Tecnología |
| :--- | :--- | :--- | :--- |
| **4.** | **Event Store** (`GormEventStore`) | **Guarda el Evento:** Persiste el `ProductCreated` en la tabla `events`. | PostgreSQL + GORM |
| **5.** | **Control de Concurrencia** | **Asegura Atomicidad:** Utiliza el índice `UNIQUE(aggregate_id, version)` para forzar la **Concurrencia Optimista** (PostgreSQL/GORM). | PostgreSQL |
| **6.** | **Event Publisher** (`KafkaPublisher`) | **Publica el Evento:** Envía el `ProductCreated` al tópico `domain_events`. | Kafka |
| **7.** | **Respuesta al Cliente** | Responde **`202 Accepted`** (Aceptado para procesamiento asíncrono). | Echo |

---

## 2. 🔎 Flujo de Consulta (Lectura)

Este flujo es **asíncrono** y **optimizado para la velocidad**, utilizando la vista desnormalizada de **MongoDB**.

### A. Proyección (Construcción del Modelo de Lectura)
| Paso | Componente | Acción Clave | Tecnología |
| :--- | :--- | :--- | :--- |
| **8.** | **Kafka Consumer** (Worker Separado) | **Consume el Evento:** Lee el evento `ProductCreated` del tópico. | Kafka |
| **9.** | **Projector** (Lógica de Proyección) | **Transforma la Data:** Traduce el evento a la estructura optimizada para lectura (`ProductDTO`). | Aplicación (Go) |
| **10.** | **Read Model Repository** | **Almacena la Vista:** Inserta o actualiza el documento en la colección `products_view`. | MongoDB |

### B. Ejecución de la Consulta
| Paso | Componente | Acción Clave | Tecnología |
| :--- | :--- | :--- | :--- |
| **11.** | **API REST (Echo Handler)** | Recibe el `GET /products/{id}`. | Echo |
| **12.** | **Query Handler** (`GetProductHandler`) | Orquesta la consulta. | Aplicación (Go) |
| **13.** | **Read Model Repository** | **Consulta Directa:** Recupera el documento por ID de la vista. | MongoDB |
| **14.** | **Respuesta al Cliente** | Devuelve el `ProductDTO` en JSON. | Echo |

---

### Resumen del Desacoplamiento

El sistema está fuertemente desacoplado:

* El **Command Side** solo habla con **PostgreSQL** y **Kafka**.
* El **Query Side** solo habla con **MongoDB**.
* **Kafka** actúa como el puente de garantía entre las dos responsabilidades.

La base de datos de lectura (el **Read Model** o **Query Side**) se actualiza de forma **asíncrona** a través del flujo de eventos, un proceso conocido como **Proyección** o **Event Handling**.

La clave es que la base de datos de lectura **nunca consulta directamente** a la base de datos de eventos; solo reacciona a los eventos que se publican.

---

## 🔁 Flujo de Actualización del Modelo de Lectura (Proyección)

Este proceso se realiza mediante un servicio o *worker* que actúa como **Consumidor** de eventos, ajeno a la API REST.

### 1. 📢 Publicación del Evento (Lado de Escritura)

Cuando un **Comando** (ej., `CreateProductCommand`) se ejecuta exitosamente:

* El Agregado (`Product`) genera un evento (`ProductCreated`).
* El **Event Store** (`PostgreSQL` vía GORM) guarda este evento de forma transaccional.
* El **Event Publisher** (`KafkaPublisher`) toma el evento recién guardado y lo envía al *Message Broker* (**Kafka**).

### 2. 👂 Consumo y Deserialización (Lado Asíncrono)

Un servicio o *worker* (el **Consumer** de Kafka), que está configurado para escuchar el tópico de eventos (`domain_events`):

* Recibe el mensaje de Kafka que contiene el evento (`ProductCreated`).
* **Deserializa** el *payload* (JSON) de vuelta a su estructura Go original.
* Pasa el evento deserializado a un **Proyector** (o *Event Handler*).

### 3. 🔄 Proyección y Almacenamiento

El **Proyector** es la lógica que sabe cómo el evento debe modificar el modelo de lectura:

* El Proyector recibe, por ejemplo, el evento `ProductCreated`.
* Sabe que este evento requiere crear un nuevo documento en la colección `products_view` de **MongoDB**.
* Utiliza la información del evento (ID,