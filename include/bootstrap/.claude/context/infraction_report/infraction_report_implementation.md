# Contexto: Implementação InfractionReport - Passo 7

## Status do Projeto

### ✅ Completado:

- **Passo 1-4**: Análise completa (Claims, SDK, Monitor, Documentação)
- **Passo 6**: Implementação orchestration-worker (COMPLETO)

### 🔄 Em Progresso:

- **Passo 7**: Implementação orchestration-monitor

### ⏳ Pendente:

- **Passo 8**: Testes unitários

---

## Implementação orchestration-worker (COMPLETO)

### Arquivos Criados (27 arquivos):

#### 1. Ports e Interfaces (1 arquivo)

- `apps/orchestration-worker/application/ports/infraction_report.go`
  - Interface `InfractionReportService` com 4 métodos

#### 2. Activities (6 arquivos)

- `infrastructure/temporal/activities/infraction_reports/infraction_report_activity.go` (base)
- `infrastructure/temporal/activities/infraction_reports/create_activity.go`
- `infrastructure/temporal/activities/infraction_reports/update_activity.go`
- `infrastructure/temporal/activities/infraction_reports/cancel_activity.go`
- `infrastructure/temporal/activities/infraction_reports/close_activity.go`
- `infrastructure/temporal/activities/infraction_reports/get_infraction_report_activity.go`

#### 3. Workflows (6 arquivos)

- `infrastructure/temporal/workflows/infraction_reports/shared.go`
- `infrastructure/temporal/workflows/infraction_reports/create_workflow.go`
- `infrastructure/temporal/workflows/infraction_reports/update_workflow.go`
- `infrastructure/temporal/workflows/infraction_reports/cancel_workflow.go`
- `infrastructure/temporal/workflows/infraction_reports/close_workflow.go`
- `infrastructure/temporal/workflows/infraction_reports/monitor_status_workflow.go`

#### 4. Services (1 arquivo)

- `infrastructure/temporal/services/infraction_report_service.go`

#### 5. UseCases (5 arquivos)

- `application/usecases/infraction_report/application.go`
- `application/usecases/infraction_report/create_infraction_report.go`
- `application/usecases/infraction_report/update_infraction_report.go`
- `application/usecases/infraction_report/cancel_infraction_report.go`
- `application/usecases/infraction_report/close_infraction_report.go`

#### 6. Handlers Pulsar (5 arquivos)

- `handlers/pulsar/infraction_report/infraction_report_handler.go`
- `handlers/pulsar/infraction_report/create_infraction_report_handler.go`
- `handlers/pulsar/infraction_report/update_infraction_report_handler.go`
- `handlers/pulsar/infraction_report/cancel_infraction_report_handler.go`
- `handlers/pulsar/infraction_report/close_infraction_report_handler.go`

#### 7. Infrastructure gRPC (1 arquivo)

- `infrastructure/grpc/infraction_report_client.go`

### Arquivos Modificados (4 arquivos):

1. **`infrastructure/grpc/gateway.go`**

   - Adicionado `InfractionReportsClient` field
   - Inicialização no constructor

2. **`setup/config.go`**

   - 4 novos campos Pulsar topic:
     - `PulsarTopicCreateInfractionReport`
     - `PulsarTopicUpdateInfractionReport`
     - `PulsarTopicCancelInfractionReport`
     - `PulsarTopicCloseInfractionReport`

3. **`setup/pulsar.go`**

   - Adicionado `infractionReportHandler` em `PulsarHandlers`
   - Subscrição nos 4 novos tópicos
   - 4 rotas `OnMessage` configuradas

4. **`setup/temporal.go`**

   - 5 workflows registrados
   - 5 activities registradas
   - Import de packages InfractionReports

5. **`setup/setup.go`**
   - Service `InfractionReportService` criado
   - Application `infractionReportApp` criado
   - Handler `infractionReportHandler` criado e injetado

---

## Próximo Passo: orchestration-monitor

### Contexto do Monitor:

O **orchestration-monitor** tem a função de:

1. **Listar** periodicamente infraction reports onde `IsReporter=true`
2. **Fechar automaticamente** aqueles com status OPEN que atingiram critério de fechamento
3. Usar **cursor Redis** para controle de paginação
4. Usar **ContinueAsNew** para workflows de longa duração

### Padrão de Referência (Claims):

#### Activities:

- `list_claims_activity.go`: Lista claims com paginação
- `complete_claim_activity.go`: Completa/fecha claim

#### Workflows:

- `complete_claim_workflow.go`: Workflow child para completar um claim específico
- `list_claims_workflow.go`: Workflow principal com loop + ContinueAsNew

#### Setup:

- `monitor_starter_process.go`: Inicia workflow de listagem periodicamente
- `temporal.go`: Registra workflows/activities do monitor

### Diferenças InfractionReport vs Claims:

| Aspecto            | Claims          | InfractionReport            |
| ------------------ | --------------- | --------------------------- |
| Ação de fechamento | Complete        | Close                       |
| Filtro de listagem | `isDonor=true`  | `IsReporter=true`           |
| Status final       | COMPLETED       | CLOSED                      |
| Cursor key         | `claims:cursor` | `infraction_reports:cursor` |

---

## Plano Passo 7: Implementação orchestration-monitor

### 7.1 Activities (3 arquivos)

- `infraction_report_activity.go` (base)
- `list_infraction_reports_activity.go`
- `close_infraction_report_activity.go`

### 7.2 Workflows (2 arquivos)

- `close_infraction_report_workflow.go` (child)
- `list_infraction_reports_workflow.go` (main com ContinueAsNew)

### 7.3 Infrastructure gRPC (1-2 arquivos)

- Atualizar `gateway.go` (adicionar InfractionReportsClient se não existir)
- Criar `infraction_report_client.go` (se não existir)

### 7.4 Setup (3 arquivos)

- Atualizar `config.go` (ReporterParticipant, CursorKeyInfractionReport)
- Atualizar `temporal.go` (registrar workflows/activities)
- Atualizar `monitor_starter_process.go` (adicionar ListInfractionReportsWorkflow)

---

## Informações Técnicas Importantes

### SDK Contracts (InfractionReport):

```go
// Listagem
type ListInfractionReportsRequest struct {
    Headers ListInfractionReportsRequestHeaders
    Filters ListInfractionReportsFilters
}

type ListInfractionReportsFilters struct {
    IsReporter        *bool
    InfractionReportID *string
    Status            *string
    Cursor            *string
}

// Fechamento
type CloseInfractionReportRequest struct {
    InfractionReportID string
    Headers           CloseInfractionReportRequestHeaders
}
```

### Workflow Pattern (List):

```go
1. Buscar cursor do Redis
2. Loop de polling:
   a. ExecuteActivity(ListInfractionReports, cursor)
   b. Para cada item OPEN:
      - ExecuteChildWorkflow(CloseInfractionReportWorkflow)
   c. Atualizar cursor no Redis
   d. Sleep(pollingInterval)
3. Se executando por muito tempo: ContinueAsNew
```

### Workflow Pattern (Close):

```go
1. ExecuteActivity(CloseInfractionReport)
2. ExecuteActivity(CreateResponseCache)
3. ExecuteActivity(PublishCoreEvents)
4. ExecuteActivity(PublishDictEvents)
```

---

## Convenções de Nomenclatura

### Package names:

- `infractionreports` (workflows/activities)

### Import aliases:

```go
activitiesInfractionReports "path/to/activities/infraction_reports"
workflowsInfractionReports "path/to/workflows/infraction_reports"
infractionreport "sdk/bacen/infraction_report"
infractionreportmapper "sdk/mappers/infraction_report"
pb "sdk/grpc/infraction_report"
```

### Activity/Workflow names:

- `ListInfractionReportsActivityName`
- `CloseInfractionReportActivityName`
- `CloseInfractionReportWorkflow`
- `ListInfractionReportsWorkflow`

---

## Validações SDK

### Actions disponíveis:

- ✅ `pkg.ActionCreateInfractionReport`
- ✅ `pkg.ActionUpdateInformationInfractionReport`
- ✅ `pkg.ActionCancelInfractionReport`
- ✅ `pkg.ActionCloseInfractionReport`

### Status InfractionReport:

- `OPEN`
- `ACKNOWLEDGED`
- `CLOSED`
- `CANCELLED`

## Visão Geral

Implementação de workflows temporais para gerenciamento de notificações de infração (Infraction Reports) nos apps `orchestration-worker` e `orchestration-monitor`, seguindo os padrões já estabelecidos para Claims.

## Arquitetura dos Apps

### orchestration-worker

**Responsabilidade**: Consumir eventos Pulsar e executar workflows temporais que orquestram operações via gRPC, cache e publicação de eventos.

**Estrutura de pastas**:

```
handlers/pulsar/         # Handlers para mensagens Pulsar
application/usecases/    # Casos de uso (delegam para services)
infrastructure/
  temporal/
    workflows/           # Workflows temporais
    activities/          # Activities (gRPC, cache, events)
    services/            # Services que executam workflows
  grpc/                  # Clientes gRPC
  pulsar/                # Publishers Pulsar
setup/                   # Configuração e injeção de dependências
```

**Fluxo de execução**:

1. Handler Pulsar → UseCase → Service → Workflow
2. Workflow → Activities (gRPC, Cache, Events)
3. Child Workflows para monitoramento de status

### orchestration-monitor

**Responsabilidade**: Monitorar status de objetos e publicar eventos quando há mudanças.

**Estrutura**:

```
infrastructure/temporal/
  workflows/             # Workflows de listagem e processamento
  activities/            # Activities (gRPC, cache, events)
setup/
  monitor_starter_process.go  # Inicia workflows de monitoramento
```

**Fluxo**:

1. MonitorStarterProcess inicia workflow de listagem
2. Workflow lista objetos (polling)
3. Para cada objeto encontrado, executa child workflow de processamento
4. Usa ContinueAsNew para manter execução contínua

## Padrões Identificados

### Claims no orchestration-worker

#### Handlers Pulsar

- Localização: `handlers/pulsar/claim/`
- Padrão: Decodificar mensagem → Chamar usecase
- Arquivos: `create_claim_handler.go`, `complete_claim_handler.go`, `cancel_claim_handler.go`, `confirm_claim_handler.go`

#### UseCases

- Localização: `application/usecases/claim/`
- Padrão: Delegar para ClaimService (que executa workflows Temporal)
- Application struct: Injeta dependências (service, observability)

#### Workflows

Localização: `infrastructure/temporal/workflows/claims/`

**CreateClaimWorkflow**:

1. ExecuteActivity: CreateClaimActivity (gRPC)
2. ExecuteCacheActivity
3. ExecuteCoreEventsPublishActivity
4. ExecuteDictEventsPublishActivity
5. StartChildWorkflow: ExpireCompletionPeriodEndWorkflow
6. StartChildWorkflow: MonitorClaimStatusWorkflow

**CompleteClaimWorkflow / CancelClaimWorkflow**:

1. ExecuteActivity: gRPC
2. ExecuteCacheActivity
3. ExecuteCoreEventsPublishActivity
4. ExecuteDictEventsPublishActivity

**MonitorClaimStatusWorkflow**:

- Polling com `Sleep(30s)` e verificação de status
- Usa `ContinueAsNew` quando atinge timeout
- Publica eventos quando status muda para CANCELLED ou CONFIRMED

#### Activities

Localização: `infrastructure/temporal/activities/claims/`

- **GRPC Activities**: `create_activity.go`, `complete_activity.go`, `cancel_activity.go`, `get_claim_activity.go`
  - Padrão: Chamar grpcGateway → Tratar erros (retryable vs non-retryable) → Mapear resposta
- **Cache Activities**: `activities/cache/set_cache_activity.go`
- **Events Activities**: `activities/events/core_events_activity.go`, `dict_events_activity.go`

#### Services

- `ClaimService`: Implementa interface `ports.ClaimService`
- Executa workflows via `temporalClient.ExecuteWorkflow()`
- Usa `WorkflowIDReusePolicy: ALLOW_DUPLICATE_FAILED_ONLY`

### Claims no orchestration-monitor

#### Workflows

**AcknowledgeClaimWorkflow**:

1. ExecuteActivity: AcknowledgeClaimActivity
2. ExecuteCoreEventsPublishActivity
3. ExecuteDictEventsPublishActivity

**ListClaimsWorkflow**:

1. Busca cursor no cache
2. Loop polling (10s interval, 20s max duration)
3. Para cada claim OPEN: ExecuteChildWorkflow(AcknowledgeClaimWorkflow)
4. Atualiza cursor no cache
5. ContinueAsNew para reiniciar ciclo

#### MonitorStarterProcess

- Garante apenas 1 monitor ativo
- Cleanup de workflows anteriores (terminate + delete)
- Inicia `ListClaimsWorkflow` com `TERMINATE_IF_RUNNING`

## Contratos SDK (InfractionReport)

### Requests/Responses disponíveis:

1. **CreateInfractionReportRequest/Response**

   - Campos: `InfractionReport`, `Participant`
   - Response: `InfractionReport` (ExtendedInfractionReport)

2. **UpdateInformationInfractionReportRequest/Response**

   - Campos: `InfractionReportID`, `Participant`, `SituationType`, `ContactInformation`, `ReportDetails`

3. **GetInfractionReportRequest/Response**

   - Campos: `InfractionReportID`

4. **CancelInfractionReportRequest/Response**

   - Campos: `InfractionReportID`, `Participant`

5. **CloseInfractionReportRequest/Response**

   - Campos: `InfractionReportID`, `Participant`, `AnalysisResult`, `FraudType`, `AnalysisDetails`

6. **ListInfractionReportsRequest/Response**

   - Campos: `Participant`, `IsReporter`, `IsCounterparty`, `Status[]`, `ModifiedAfter`, etc.

7. **AcknowledgeInfractionReportRequest/Response** (para monitor)
   - Campos: `InfractionReportID`, `Participant`

### Status de InfractionReport

- OPEN
- ACKNOWLEDGED
- CLOSED
- CANCELLED

## Configurações

### orchestration-worker

ENV vars necessárias (já existem):

- `PULSAR_TOPIC_CREATE_INFRACTION_REPORT`
- `PULSAR_TOPIC_UPDATE_INFRACTION_REPORT`
- `PULSAR_TOPIC_CANCEL_INFRACTION_REPORT`
- `PULSAR_TOPIC_CLOSE_INFRACTION_REPORT`

### orchestration-monitor

ENV vars necessárias (criar):

- `REPORTER_PARTICIPANT` (número do participante reporter)
- `CURSOR_KEY_INFRACTION_REPORT` (chave para cursor de listagem)

## Interfaces e Injeção de Dependências

### orchestration-worker

- **Port**: `InfractionReportService` interface
- **Implementation**: `InfractionReportService` struct com Temporal client
- **Injection**: Setup cria service e injeta em Application

### orchestration-monitor

- Usa mesma estrutura de activities e eventos
- Não precisa de service layer (executa workflows direto no starter)

## Testes Unitários

### Padrão identificado:

- Localização: `tests/unit/infrastructure/temporal/`
- Estrutura: `activities/`, `workflows/`, `utils/`
- Ferramentas: `testify/suite`, mocks de Temporal
- Helper: `helper_tests.go` com funções auxiliares

### Cobertura necessária:

- Todos workflows: casos de sucesso e erro
- Todas activities: validação de erros retryable/non-retryable
- Utilities: funções helper

## Observações Importantes

1. **Actions do pkg**: Usar constantes do SDK como `pkg.ActionCreateInfractionReport`
2. **Error handling**: Distinguir entre erros retryable (transient) e non-retryable (business)
3. **ContinueAsNew**: Usar para workflows de longa duração
4. **ParentClosePolicy**: `ABANDON` para child workflows que devem continuar após parent terminar
5. **WorkflowID**: Usar padrão `{action}:{objectID}` para idempotência
6. **Cursor**: Usar cache para manter estado de listagem entre execuções
