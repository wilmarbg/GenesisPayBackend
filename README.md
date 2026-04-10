# GenesisPay Backend

GenesisPay Backend es un sistema robusto, estructurado en arquitectura de microservicios escritos en Go, diseñado para manejar autenticación, clientes, comercios y pagos de forma modular, asegurando alta disponibilidad y escalabilidad.

## 1. Diagrama de Arquitectura

```text
Flutter App → Nginx (Puerto 8080 - API Gateway)
                 │
                 ├──→ Auth Service      (:8001)
                 ├──→ Clients Service   (:8002)
                 ├──→ Merchants Service (:8003)
                 └──→ Payments Service  (:8004)
                      │
                      └──→ Base de Datos (PostgreSQL)
```

## 2. Tecnologías Usadas

- **Lenguaje:** Go (Golang)
- **Framework Web:** Gin Web Framework
- **ORM:** GORM
- **Base de Datos:** PostgreSQL
- **Autenticación:** JWT (JSON Web Tokens)
- **Cifrado de Datos Sensibles:** AES-256-GCM
- **Cifrado de Contraseñas:** bcrypt

## 3. Requisitos Previos

- **Go:** 1.22 o superior
- **PostgreSQL:** 16 o superior
- **Docker y Docker Compose** (para infraestructura en contenedores)

## 4. Instalación Paso a Paso

1. Clonar el repositorio:
   ```bash
   git clone https://github.com/wilmarbg/GenesisPayBackend.git
   cd GenesisPayBackend
   ```
2. Descargar dependencias:
   ```bash
   go mod tidy
   ```
3. Configurar variables de entorno (ver sección Variables de Entorno).
4. Asegurarse de tener PostgreSQL en ejecución o usar Docker Compose.

## 5. Variables de Entorno (.env)

Crea un archivo `.env` en la raíz del proyecto basado en `.env.example`:

```env
DB_HOST=172.17.0.1
DB_PORT=5432
DB_USER=genesis_user
DB_PASSWORD=secret
DB_NAME=genesis_db
DB_SSL_MODE=disable

AUTH_PORT=8001
CLIENTS_PORT=8002
MERCHANTS_PORT=8003
PAYMENTS_PORT=8004

JWT_SECRET=tu_secreto_muy_seguro_y_largo
JWT_EXPIRATION_HOURS=24
ENCRYPTION_KEY=clave_fuerte_de_32_bytes_12345678 # Debe tener exactamente 32 bytes
```

## 6. Ejecución

### En Desarrollo (Local sin Contenedores)

Para ejecutar el sistema completo localmente, necesitas abrir 4 terminales diferentes e iniciar cada microservicio:

```bash
# Terminal 1
go run cmd/auth/main.go

# Terminal 2
go run cmd/clients/main.go

# Terminal 3
go run cmd/merchants/main.go

# Terminal 4
go run cmd/payments/main.go
```

### Con Docker

Puedes levantar toda la infraestructura (incluyendo la base de datos de ser necesario) usando Docker Compose:

```bash
# Construir y levantar contenedores
docker-compose up --build

# Ver logs de un servicio (ejemplo auth)
docker-compose logs -f auth

# Detener los contenedores
docker-compose down
```

## 7. Tabla de Endpoints

A continuación se detalla la tabla completa de los endpoints implementados:

### Auth (Puerto :8001)
| Método | Endpoint | Descripción | Perfil/Rol |
|---|---|---|---|
| POST | `/api/v1/auth/register` | Registrar un nuevo usuario | Público |
| POST | `/api/v1/auth/login` | Iniciar sesión y obtener token | Público |
| GET | `/api/v1/auth/me` | Obtener datos del perfil actual | Autenticado |

### Clients (Puerto :8002)
| Método | Endpoint | Descripción | Perfil/Rol |
|---|---|---|---|
| POST | `/api/v1/clients` | Crear un cliente (CRUD) | Autenticado |
| GET | `/api/v1/clients` | Listar todos los clientes (CRUD) | Admin |
| GET | `/api/v1/clients/:id` | Obtener cliente por ID (CRUD) | Admin/Dueño |
| PUT | `/api/v1/clients/:id` | Actualizar cliente (CRUD) | Admin/Dueño |
| DELETE | `/api/v1/clients/:id` | Desactivar cliente (CRUD) | Admin |
| GET | `/api/v1/clients/me` | Obtener mi perfil como cliente | Autenticado |
| PATCH | `/api/v1/clients/:id/status` | Actualizar el estado del cliente | Admin |
| PATCH | `/api/v1/clients/:id/activate` | Activar un cliente específico | Admin |

### Merchants (Puerto :8003)
| Método | Endpoint | Descripción | Perfil/Rol |
|---|---|---|---|
| POST | `/api/v1/merchants` | Crear comercio (CRUD) | Admin |
| GET | `/api/v1/merchants` | Listar comercios (CRUD) | Público |
| GET | `/api/v1/merchants/:id` | Detalle comercio (CRUD) | Público |
| PUT | `/api/v1/merchants/:id` | Actualizar comercio (CRUD) | Admin/Dueño |
| DELETE | `/api/v1/merchants/:id` | Desactivar comercio (CRUD) | Admin |
| GET | `/api/v1/merchants/me` | Obtener mi comercio actual | Autenticado (Dueño) |
| GET | `/api/v1/merchants/:id/products` | Ver productos del comercio | Público |
| POST | `/api/v1/merchants/:id/products` | Crear un producto en el comercio | Dueño |
| PUT | `/api/v1/merchants/:id/products/:pid`| Actualizar un producto | Dueño |
| DELETE | `/api/v1/merchants/:id/products/:pid`| Eliminar un producto | Admin/Dueño |

### Payments (Puerto :8004)
| Método | Endpoint | Descripción | Perfil/Rol |
|---|---|---|---|
| POST | `/api/v1/cards` | Emitir/registrar nueva tarjeta | Admin |
| GET | `/api/v1/cards/me` | Listar las tarjetas de mi perfil | Cliente |
| GET | `/api/v1/cards/:id/balance` | Obtener saldo de tarjeta | Cliente/Admin |
| PATCH | `/api/v1/cards/:id/freeze` | Congelar la tarjeta | Admin |
| PATCH | `/api/v1/cards/:id/cancel` | Cancelar la tarjeta | Admin |
| POST | `/api/v1/payments` | Procesar un pago/compra | Cliente |
| GET | `/api/v1/payments` | Listar todos los pagos | Autenticado |
| GET | `/api/v1/payments/:id` | Detalle de un pago | Autenticado |
| POST | `/api/v1/payments/:id/refund` | Reembolsar transacción | Admin |

## 8. Reglas de Negocio

- **Estados del Cliente y transiciones válidas:** Un cliente pasa por diferentes estados controlados (por ejemplo, `PENDIENTE` -> `ACTIVO` -> `SUSPENDIDO`). Solo se permiten transiciones lógicas predefinidas validadas estrictamente por la lógica de negocio.
- **Comercios y estado activo:** Un comercio nace *activo* inmediatamente al crearse sin requerir una fase de aprobación, para agilizar la publicación del catálogo y evitar fricción inicial.
- **Creación de Administrador:** El rol de administrador (`admin`) no se expone a través de la API pública para evitar escalada de privilegios; los usuarios administradores se crean o promueven únicamente de forma directa mediante la base de datos de manera manual (DB level).

## 9. Diseño de Base de Datos

- **4 Schemas Independientes:**
  - `aut` (Autenticación)
  - `cli` (Clientes)
  - `com` (Comercios)
  - `pag` (Pagos)
- **Aislamiento por Schema:** Cada microservicio gestiona su propio schema garantizando el aislamiento de datos y previniendo acoplamiento de base de datos.
- **Excepciones de Aislamiento:** Aunque los datos están separados, las lógicas de *Payments* y *Clients* en ciertas ocasiones pueden acceder a múltiples schemas de solo lectura para cruzar información crítica en procesamientos o validaciones relacionales complejas.

## 10. Arquitectura del Código

La aplicación sigue principios de arquitectura limpia de tres niveles:
1. **Handler (Controlador):** Gestiona la recepción de la petición HTTP, validación estructural básica (JSON bindings) y envía las respuestas (JSON HTTP status).
2. **Service (Lógica de Negocio):** Otorga el cerebro a las operaciones; ejecuta reglas de dominio (ej. validaciones de negocio, permisos lógicos), procesa los datos y gestiona los errores de negocio.
3. **Repository (Persistencia):** Encapsula todas las consultas, interacciones y transacciones a la base de datos aislando GORM del resto del código. (Flujo: `Handler -> Service -> Repository`).

## 11. Auditoría

- **Tabla de Registros:** Toda la actividad transaccional crítica y cambios de estado se almacenan en la tabla `pag.audit_logs`.
- **Acciones Registradas:** Se guarda un json con el impacto en base de datos indicando `user_id`, `entity_type`, la `action` y el timestamp, garantizando rastreabilidad total.
- **PCI DSS Básico:** Evita el almacenamiento en texto plano del número de la tarjeta de crédito y retiene información requerida de acceso para el cumplimiento del estándar PCI DSS (Payment Card Industry Data Security Standard) de forma básica.

## 12. Seguridad Implementada

- **JWT con algoritmo HS256:** Para asegurar una autenticación inmutable y veloz de forma stateless, verificando firmas secretas.
- **Hashing de Contraseñas:** Se usa `bcrypt` con un `cost` de 10 para garantizar resiliencia computacional ante ataques de fuerza bruta.
- **Cifrado Fuerte:** Para proteger los datos reales (PAN) de las tarjetas, se utiliza encriptación simétrica bidireccional `AES-256-GCM`.
- **Tokenización de Tarjetas:** Se expone únicamente un `card_token` público en formato UUID, manteniendo el PAN oculto en reportes y respuestas.
- **Sentencias Preparadas:** Protegido contra Inyecciones SQL usando las variables bind y placeholders de `GORM` (`?`).
- **Validación de Roles:** Middlewares dedicados en cada endpoint que verifican tanto validez del token como niveles exigidos de autorización (Ej: *Admin, Dueño, Cliente*).
- **Validación de Transiciones de Estado:** Bloqueos a nivel backend contra mutaciones no autorizadas en el ciclo de vida de una entidad.

## 13. Estructura Completa del Proyecto

```text
/
├── cmd/               # Puntos de entrada para cada servicio
│   ├── auth/          # Main de Auth
│   ├── clients/       # Main de Clients
│   ├── merchants/     # Main de Merchants
│   └── payments/      # Main de Payments
├── docker-compose.yml # Composición de servicios y base de datos
├── gateway/           # Nginx reverse proxy routing y balanceo (.conf)
├── internal/          # Lógica de dominio acotada por servicio
│   ├── auth/          # Handlers, Services, Repositories de Auth
│   ├── clients/       # Handlers, Services, Repositories de Clients
│   ├── config/        # Funciones globales y configuración (enrutador, DB)
│   ├── database/      # Conexión principal PostgreSQL
│   ├── merchants/     # Handlers, Services, Repositories de Merchants
│   └── payments/      # Handlers, Services, Repositories de Payments
├── middleware/        # Interceptores compartidos (CORS, Roles, JWT)
└── Dockerfile.*       # Recetas Docker optimizadas (Auth, Clients, Merchants, Payments)
```

## 14. Cómo Probar la API

A continuación unos ejemplos de cómo puedes disparar peticiones usando `curl` (puedes importarlos o replicarlos en Postman):

**Ejemplo de Registro (Auth):**
```bash
curl -X POST http://localhost:8001/api/v1/auth/register \
     -H "Content-Type: application/json" \
     -d '{
           "email": "nuevo@correo.com",
           "password": "miPasswordSeguro123",
           "name": "Juan Perez"
         }'
```

**Ejemplo de Login (Auth) para obtener Token:**
```bash
curl -X POST http://localhost:8001/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{
           "email": "nuevo@correo.com",
           "password": "miPasswordSeguro123"
         }'
```

**Ejemplo de usar Endpoint Protegido con Token (Clients):**
```bash
curl -X GET http://localhost:8002/api/v1/clients/me \
     -H "Authorization: Bearer <TU_TOKEN_AQUI>"
```
