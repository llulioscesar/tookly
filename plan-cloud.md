# Canonical Core y Tookly Cloud — Arquitectura de Producto y Diseño Técnico

**Autor:** Julio Cesar · StartCodex
**Fecha:** Marzo 2026
**Versión:** 3.1
**Clasificación:** Documento interno — no distribuir

---

## 1. Estado actual de Tookly

Tookly es un monolito Go 1.26 con SvelteKit embebido en un solo binario. PostgreSQL con sqlx, stdlib HTTP router, frontend compilado y servido desde el mismo proceso. La arquitectura sigue el patrón de un paquete por dominio, funciones libres con dependencias explícitas, SQL directo y sin abstracciones preventivas.

Entregado: workspaces, proyectos, issue types, boards, statuses, issues CRUD, MoveIssue API, login, i18n (EN/ES).

En progreso (Phase 1): auth con sesiones server-side, membership enforcement, board UI, issue detail.

Planeado: sprints y backlog (Phase 2), documentación como dominio de primera clase con links docs↔work items (Phase 3), templates cross-industry (Phase 4), automations y reporting (Phase 5), AI assistant y MCP (Phase 6).

El roadmap público establece que la planificación basada en documentación entra en Phase 3 con links manuales, y que la inferencia asistida llega después. Este documento define la arquitectura privada que opera por encima de esa base.

---

## 2. Producto

El producto se compone de tres piezas con ownership, licencia y deploy separados.

### Tookly OSS

Workflow platform source-available bajo BSL 1.1. Self-hostable, un solo binario. El scope de producto incluye workspaces, proyectos, issues, boards, documentación, ADRs, backlog y sprints — entregados progresivamente según el roadmap público. Puede exponer workflow assistant, configuración provider-agnostic de IA, propuestas de origen humano o IA, documentación asistida y MCP para sistemas externos. No expone el motor semántico privado.

### Tookly Cloud

SaaS administrado. Reutiliza Tookly OSS como base y agrega integraciones privadas, observabilidad, budget controls, governance flows y el Canonical Core. Cuando un usuario crea un documento, el sistema evalúa alineación y detecta posibles inconsistencias con el backlog. Cuando cambia una decisión, indica qué queda stale. Cuando solicita una propuesta, el LLM genera un changeset estructurado.

Tookly Cloud no cobra por gestionar trabajo. Cobra por mantener coherencia entre lo que el equipo piensa, documenta, decide y construye.

### Canonical Core

Servicio privado, cerrado, con API propia. Nunca se distribuye. Nunca vive en el mismo binario que Tookly. Se comunica exclusivamente por red. Knowledge graph tipado, changesets controlados, reglas de alineación, análisis de impacto, propuestas LLM estructuradas y proyecciones documentales.

---

## 3. Topología del código

```
tookly-oss/        Repo público, source-available (BSL 1.1)
tookly-cloud/      Repo privado, construye sobre OSS
canonical-core/    Repo privado, motor semántico
```

Dirección de dependencias:

- Cloud depende de OSS.
- Cloud depende de Canonical Core.
- OSS no depende de Cloud.
- OSS no depende de Canonical Core.
- Canonical Core no depende de internals de OSS.

Cloud es una capa privada que reusa la base pública con su propio código, deployment y operaciones. No es un feature flag dentro del repo OSS.

---

## 4. Licencia

Tookly OSS está bajo BSL 1.1.

| Parámetro | Valor |
|---|---|
| **Additional Use Grant** | Self-hosting para uso interno propio; no competing SaaS |
| **Change Date** | 4 años desde la fecha de release de cada versión |
| **Change License** | Apache 2.0 |

La AGPL obligaría a liberar el código fuente de cualquier versión modificada ofrecida como servicio por red, lo cual destruiría la ventaja competitiva de Cloud. GPLv3 no obliga a liberar código por red, pero dejaría a Tookly desprotegido frente a competidores que monten un servicio hospedado. BSL 1.1 mantiene el código visible, permite self-hosting para uso interno y bloquea el uso competitivo como servicio administrado sin licencia comercial.

---

## 5. Rust para el Canonical Core

Tookly seguirá siendo Go. El Canonical Core se implementa en Rust.

El canonical no es un servicio CRUD. Es un graph engine que hace traversal de relaciones en cadena, propagación de stale sobre subgrafos, ejecución de reglas sobre el grafo completo después de cada changeset, y pattern matching exhaustivo sobre estados y variantes del dominio. Los enums algebraicos de Rust fuerzan a manejar todos los casos en compilación. El ownership model elimina bugs de memoria sin GC. El costo de reescribir un engine de Go a Rust cuando ya tiene tracción es mayor que el costo de empezar en Rust.

Go y Rust no se mezclan. Son dos servicios independientes que hablan por red. El canonical no necesita iterar al ritmo de UI y auth de Tookly.

| Componente | Librería |
|---|---|
| HTTP | axum |
| gRPC | tonic |
| Postgres | sqlx |
| Serialización | serde + serde_json |
| CLI | clap |
| Async | tokio |
| Testing | built-in + proptest |
| Logging | tracing |

---

## 6. Principios de arquitectura

- Tookly OSS no depende del Canonical Core.
- Tookly Cloud integra el Canonical exclusivamente por red.
- El dominio del Canonical no depende de infraestructura.
- Qdrant nunca es fuente de verdad.
- Todo cambio al grafo entra por ChangeSet.
- El LLM nunca modifica el grafo directamente; solo propone cambios estructurados.
- La identidad del usuario se propaga hasta el canonical para auditoría y governance.
- El moat vive en el motor semántico y en el control plane privado, no en la superficie básica de assistant expuesta en OSS.

---

## 7. Arquitectura del Canonical Core

### 7.1 Topología

```
┌─────────────────────────────────────┐
│           Tookly Cloud              │
│   (Tookly base + integraciones)     │
│                                     │
│  Cuando el usuario crea/cambia      │
│  un doc, decisión o backlog item,   │
│  Cloud llama al canonical vía API   │
└──────────────┬──────────────────────┘
               │ HTTP/gRPC
               ▼
┌─────────────────────────────────────┐
│         Canonical Core              │
│      (servicio privado, cerrado)    │
│                                     │
│  Knowledge graph, changesets,       │
│  alignment rules, impact analysis,  │
│  LLM proposals, projections         │
│                                     │
│  Postgres propio + Qdrant           │
└─────────────────────────────────────┘
```

### 7.2 Interfaces

| Interfaz | Consumidor | Uso |
|---|---|---|
| **API REST/gRPC** | Tookly Cloud | Integración programática |
| **CLI** | Desarrollo, CI/CD | Testing, scripts, debug |
| **SDK/Client** | Integraciones futuras | Otros productos de la compañía |

### 7.3 Endpoints

Superficie HTTP de referencia. La interfaz gRPC expone las mismas operaciones con contratos protobuf propios.

```
POST   /v1/workspaces/{id}/changesets
GET    /v1/workspaces/{id}/graph
GET    /v1/workspaces/{id}/validation
GET    /v1/workspaces/{id}/impact/{artifact}
POST   /v1/workspaces/{id}/search/similar
GET    /v1/workspaces/{id}/projections/{type}
POST   /v1/workspaces/{id}/proposals
POST   /v1/workspaces/{id}/audit
GET    /v1/workspaces/{id}/audit/latest
```

Auth de transporte: service-to-service con mTLS o API key interna. Auth de negocio: cada request incluye identidad del usuario final y contexto del workspace para autorización y auditoría.

---

## 8. Estructura del Canonical Core

Cargo workspace. Un solo binario con servidor HTTP, CLI y migraciones.

### 8.1 Layout

```
canonical-core/
├── Cargo.toml
├── crates/
│   ├── canonical-domain/
│   ├── canonical-engine/
│   ├── canonical-rules/
│   ├── canonical-projection/
│   ├── canonical-serde/
│   └── canonical-llm/
├── src/
│   ├── main.rs
│   ├── api/
│   ├── db/
│   ├── config.rs
│   └── error.rs
├── migrations/
├── tests/
├── Dockerfile
└── README.md
```

### 8.2 Dependencias

```
canonical-domain          (leaf — solo deps de modelado: serde, uuid, chrono)
       ▲
canonical-engine          (depende de domain)
       ▲
canonical-rules           (depende de domain + engine)

canonical-projection      (depende de domain)
canonical-serde           (depende de domain)
canonical-llm             (depende de domain + serde)

src/ (binary)             (depende de todos + axum, sqlx, clap)
```

Los crates de dominio nunca dependen de infraestructura.

### 8.3 Tipos del dominio

```rust
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum ArtifactKind {
    Vision,
    BusinessGoal,
    Capability,
    BusinessRule,
    Requirement,
    UseCase,
    Constraint,
    Assumption,
    Risk,
    OpenQuestion,
    Decision,
    ArchitectureElement,
    BacklogItem,
    TestScenario,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum ArtifactStatus {
    Draft,
    Proposed,
    Reviewed,
    Approved,
    Rejected,
    Deprecated,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum RelationKind {
    DerivedFrom,
    Refines,
    TracesTo,
    Implements,
    Validates,
    DependsOn,
    Impacts,
    Justifies,
    Mitigates,
    ConflictsWith,
    Supersedes,
}

pub struct Engine {
    rules: RuleRegistry,
}
```

---

## 9. Persistencia

### 9.1 Jerarquía de verdad

1. **Modelo de dominio del Canonical Core** — verdad lógica del sistema
2. **PostgreSQL** — verdad persistida (identidad, relaciones, versiones, historia)
3. **Estado en memoria** — representación transitoria para aplicar changesets y ejecutar reglas
4. **Qdrant** — índice semántico auxiliar para retrieval y contexto LLM

### 9.2 PostgreSQL

Almacena todo lo transaccional, relacional y versionable. Multi-tenancy recomendada para MVP: row-level con `workspace_id` + RLS.

### 9.3 Qdrant

Solo embeddings y queries de similitud. Nunca fuente de verdad.

### 9.4 Audit trail

```sql
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL,
    event_type TEXT NOT NULL,
    actor TEXT NOT NULL,
    request_id UUID,
    trace_id TEXT,
    source_surface TEXT,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);
```

Sin UPDATE ni DELETE. Solo INSERT. Los campos `request_id`, `trace_id` y `source_surface` se propagan desde el contexto de identidad de Cloud (§10.1), permitiendo correlacionar cualquier entrada del audit log con la request original, la traza distribuida y la superficie de origen (UI, assistant, proposal-review, system-audit). Registra cada changeset aplicado o rechazado, cada governance gate y su resolución, cada llamada al LLM, cada cambio de status, cada conflict y su resolución, cada audit report generado.

---

## 10. Integración Cloud ↔ Canonical

Tookly Cloud reusa Tookly OSS como base y agrega una capa privada de integración.

```
tookly-cloud/
├── internal/
│   ├── ...
│   └── canonical/
│       ├── client.go
│       ├── sync.go
│       ├── hooks.go
│       ├── proposal_flow.go
│       └── audit_flow.go
```

La capa intercepta eventos del dominio de Tookly, los traduce a artefactos canónicos, llama a la API del canonical, muestra resultados en la UI y aplica governance, budget controls y auditoría. Tookly OSS no conoce el canonical.

### 10.1 Propagación de identidad

Cada request de Cloud al canonical incluye:

| Campo | Propósito |
|---|---|
| `workspace_id` | Scope del grafo |
| `user_id` | Identidad del usuario real |
| `actor_role` | Rol para autorización y filtrado |
| `request_id` | Correlación de request |
| `trace_id` | Observabilidad distribuida |
| `source_surface` | Origen: `ui`, `assistant`, `proposal-review`, `system-audit` |

El transporte autentica al servicio. La autorización y auditoría operan con identidad de usuario real. El canonical no debe confiar en `user_id`, `actor_role` ni `workspace_id` solo porque llegaron en headers. Ese contexto debe venir firmado o encapsulado por Cloud — JWT interno, signed headers o context envelope — de modo que el canonical pueda verificar que la identidad fue emitida por una instancia legítima de Cloud y no fue manipulada en tránsito.

---

## 11. Interacción con LLM

### 11.1 Contexto filtrado por rol

Enviar el grafo completo al LLM no funciona. Cada `ActorRole` define un filtro sobre el knowledge graph que determina qué artefactos y relaciones ve el LLM.

```rust
pub struct PerspectiveFilter {
    pub role: ActorRole,
    pub artifact_kinds: Vec<ArtifactKind>,
    pub relation_kinds: Vec<RelationKind>,
    pub include_neighbors: bool,
    pub max_depth: usize,
}
```

| ActorRole | Ve | No ve |
|---|---|---|
| BusinessOwner | Vision, BusinessGoal, Capability, Risk | ArchitectureElement, BacklogItem, TestScenario |
| ProductManager | BusinessGoal, Capability, Requirement, UseCase, Constraint | ArchitectureElement salvo contexto necesario |
| DomainAnalyst | Requirement, UseCase, BusinessRule, Assumption, Constraint | ArchitectureElement, BacklogItem |
| SolutionArchitect | Decision, ArchitectureElement, Constraint, Risk + Requirements aprobados | Vision y BusinessGoal salvo contexto referencial |
| QALead | Requirement (Approved), TestScenario, UseCase, Risk | Decision internals, Assumption |
| DeliveryLead | BacklogItem, Decision (Approved), Requirement (Approved) | Vision, BusinessRule, Assumption |

### 11.2 Flujo de propuesta

1. Un usuario o evento de Cloud solicita una propuesta.
2. El canonical filtra el grafo por el rol solicitado.
3. El subgrafo se serializa en TOON o JSON compacto, incluyendo goal ancestry.
4. El LLM responde con una LlmProposal serializable.
5. La LlmProposal se convierte a ChangeSet.
6. `Engine::apply_changeset()` verifica consistencia, ejecuta reglas y analiza impacto.
7. El usuario recibe el ChangeSet propuesto + ValidationReport + ImpactReport.

### 11.3 Modelo secuencial

El canonical procesa un ChangeSet a la vez. Se valida, se aplica, se analiza impacto, y la siguiente propuesta ve el grafo actualizado. Múltiples LLMs proponiendo en paralelo sobre el mismo subgrafo generan conflictos de merge que no se resuelven automáticamente con fiabilidad.

### 11.4 Límites del LLM

El LLM no escribe en el knowledge graph. No decide qué cambios se aplican. No ve el grafo completo. No resuelve conflictos. No ejecuta alignment rules. Es un generador de propuestas estructuradas con contexto acotado. El engine es el guardrail. El humano es el decisor final.

---

## 12. Capacidades operativas de Cloud

### 12.1 Goal ancestry

Cada artefacto debe tener una o más rutas válidas de trazabilidad hacia arriba a través de relaciones DerivedFrom, TracesTo o Implements, que terminen en una Vision o BusinessGoal aprobados. No todas las rutas siguen la misma secuencia — un BacklogItem puede trazar directamente a un Requirement sin pasar por Decision, o una Decision puede trazar a un Risk sin intermediarios. Lo que importa es que exista al menos una ruta válida. El engine incluye esta ancestry en el contexto de propuestas LLM y la valida con una alignment rule: todo artefacto Approved sin ruta a un objetivo de negocio aprobado es trabajo huérfano.

### 12.2 Auditoría proactiva

Un heartbeat periódico (configurable por workspace) ejecuta alignment rules sobre el grafo completo, detecta artefactos stale sin revisión, conflictos abiertos sin resolución, artefactos Approved sin ancestry y decisions Proposed sin review. Genera un AuditReport priorizado por severidad. Cloud muestra el resultado en la UI sin que el usuario tenga que recordar verificar.

### 12.3 Budget enforcement

Cada workspace tiene un presupuesto mensual de tokens LLM. Warning a 80%, bloqueo de propuestas a 100%, admin override disponible. El control pertenece al control plane de Cloud. Define tiers de pricing naturales.

### 12.4 Governance gates

Puntos donde el sistema se detiene y espera aprobación humana:

| Gate | Activación |
|---|---|
| ChangeSet con findings Critical | ValidationReport con severidad Critical |
| Cambio de status a Approved | Cualquier artefacto pasa a Approved |
| Decisión con alto impacto | ImpactReport muestra >N artefactos stale |
| Propuesta LLM que toca artefactos Approved | Changeset modifica artefactos vigentes |
| Supersedes sobre Decision Approved | Reemplazo de decisión vigente |

Los changesets no se aplican automáticamente si hay un gate activo. Quedan en estado PendingApproval hasta resolución humana.

---

## 13. Componentes privados

| Componente | Razón |
|---|---|
| Canonical Core completo | Diferenciador del SaaS |
| Alignment rules concretas | Codifican conocimiento de proceso |
| Impact analyzer | Algoritmo de propagación |
| Ranking y recommendation engines | Valor de producto |
| LLM prompts y pipeline de proposals | IP directa |
| Módulo de integración Tookly ↔ Canonical | Lógica privada de Cloud |
| Projection renderers avanzados | Calidad del producto administrado |
| Operational knowledge administrado | Diferenciador de servicio |

---

## 14. Fases de implementación

| Fase | Alcance |
|---|---|
| **0** | BSL 1.1 establecido en OSS |
| **1a** | Domain + persistencia. Tipos, Postgres, API básica, tests de integración |
| **1b** | ChangeSet + Engine. Mutación controlada, audit log |
| **2** | Validación + alignment rules. ImpactAnalyzer, stale markers, ancestry validation |
| **3** | Vector embeddings. Qdrant, embedding provider, search semántico |
| **4** | Integración Cloud. Client HTTP/gRPC, hooks, propagación de identidad, UI |
| **5** | Proyecciones + LLM. PRD, ADR, Backlog, TraceMatrix, LlmProposal → ChangeSet, TOON |
| **6** | Governance + audit + budget. Heartbeat, gates, budget enforcement, audit trail. Capacidades exclusivas del control plane de Cloud, no del OSS |

---

## 15. Decisiones abiertas

| # | Decisión | Opciones | Impacto |
|---|---|---|---|
| 1 | Modelo de embedding | OpenAI, BGE-M3, Nomic | Costo, latencia, privacidad |
| 2 | Caché del knowledge graph | En memoria por request, Valkey, event sourcing | Rendimiento, consistencia |
| 3 | Formato de proyecciones | Markdown, JSON, HTML | Consumo por frontend |
| 4 | Resolución de conflictos concurrentes | Last-write-wins, merge manual, CRDT simplificado | UX, complejidad |
| 5 | Contrato de identidad propagada | Headers firmados, JWT interno, context envelope | Auditoría y seguridad |
| 6 | Topología privada de Cloud | Repo separado, fork privado, overlay | Mantenibilidad |
| 7 | Scope de governance gates | Mínimo viable vs enterprise-first | UX y operación |
| 8 | Budget model | Por workspace, por org, por tier | Pricing y control |

---

## 16. Resumen

Tookly OSS es la workflow platform source-available. Self-hostable, un solo binario, Go + SvelteKit + Postgres.

Tookly Cloud es el SaaS administrado: la misma base con integraciones privadas, control plane, governance, budgets y el Canonical Core.

El Canonical Core es un servicio Rust independiente con API propia. Se comunica con Cloud exclusivamente por red. Persiste en Postgres propio + Qdrant. Nunca se distribuye.

La separación es por producto, repositorio, deploy y licencia.

**Tookly es el sistema de trabajo. El canonical es el cerebro semántico que solo vive en Cloud.**