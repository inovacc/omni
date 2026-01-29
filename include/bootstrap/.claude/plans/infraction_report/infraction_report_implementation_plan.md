# Plano de Implementação: Infraction Report Worker & Monitor

## Status Geral

- [x] Análise completa (Passo 1-4)
- [x] Implementação orchestration-worker (Passo 6)
- [x] Implementação orchestration-monitor (Passo 7)
- [x] Testes unitários (Passo 8)
- [ ] Validação final e commit (Passo 9)

---

## Passo 1: Análise de Claim no orchestration-worker

**Status**: ✅ COMPLETO

### Analisado:

- [x] Handlers Pulsar (`handlers/pulsar/claim/`)
- [x] UseCases (`application/usecases/claim/`)
- [x] Workflows (`infrastructure/temporal/workflows/claims/`)
- [x] Activities (`infrastructure/temporal/activities/claims/`)
- [x] Services (`infrastructure/temporal/services/claim_service.go`)
- [x] Setup e configurações (`setup/`)

---

## Passo 2: Análise SDK - Contratos InfractionReport

**Status**: ✅ COMPLETO

### Contratos identificados:

- [x] CreateInfractionReportRequest/Response
- [x] UpdateInformationInfractionReportRequest/Response
- [x] GetInfractionReportRequest/Response
- [x] CancelInfractionReportRequest/Response
- [x] CloseInfractionReportRequest/Response
- [x] ListInfractionReportsRequest/Response
- [x] AcknowledgeInfractionReportRequest/Response

---

## Passo 3: Análise de Claim no orchestration-monitor

**Status**: ✅ COMPLETO

### Analisado:

- [x] Workflows (`infrastructure/temporal/workflows/claims/`)
- [x] Activities (`infrastructure/temporal/activities/claims/`)
- [x] MonitorStarterProcess (`setup/monitor_starter_process.go`)
- [x] Setup Temporal (`setup/temporal.go`)

---

## Passo 4: Documentação de Contexto

**Status**: ✅ COMPLETO

- [x] Criado `.claude/context/infraction_report_implementation.md`
- [x] Criado `.claude/plans/infraction_report_implementation_plan.md`

---

## Passo 5: Confirmação antes de implementar

**Status**: ✅ COMPLETO

---

## Passo 6: Implementação orchestration-worker

### 6.1 - Ports e Interfaces

**Status**: ✅ COMPLETO

#### Tarefas:

- [x] Criar `application/ports/infraction_report.go`
  - Interface `InfractionReportService` com métodos:
    - `CreateInfractionReport(ctx, requestID, request)`
    - `UpdateInfractionReport(ctx, requestID, request)`
    - `CancelInfractionReport(ctx, requestID, request)`
    - `CloseInfractionReport(ctx, requestID, request)`

### 6.2 - Activities

**Status**: ✅ COMPLETO

#### Tarefas:

- [x] Criar `infrastructure/temporal/activities/infraction_reports/infraction_report_activity.go`

  - Struct `Activity` com `grpcGateway`
  - Constructor `NewActivity(grpcGateway)`

- [x] Criar `infrastructure/temporal/activities/infraction_reports/create_activity.go`

  - Activity: `CreateInfractionReportActivity`
  - Const: `CreateInfractionReportActivityName`

- [x] Criar `infrastructure/temporal/activities/infraction_reports/update_activity.go`

  - Activity: `UpdateInfractionReportActivity`
  - Const: `UpdateInfractionReportActivityName`

- [x] Criar `infrastructure/temporal/activities/infraction_reports/cancel_activity.go`

  - Activity: `CancelInfractionReportActivity`
  - Const: `CancelInfractionReportActivityName`

- [x] Criar `infrastructure/temporal/activities/infraction_reports/close_activity.go`

  - Activity: `CloseInfractionReportActivity`
  - Const: `CloseInfractionReportActivityName`

- [x] Criar `infrastructure/temporal/activities/infraction_reports/get_infraction_report_activity.go`
  - Activity: `GetInfractionReportActivity`
  - Const: `GetInfractionReportActivityName`

### 6.3 - Workflows

**Status**: ✅ COMPLETO

#### Tarefas:

- [x] Criar `infrastructure/temporal/workflows/infraction_reports/shared.go`

  - Funções helper compartilhadas
  - Constantes de workflow names

- [x] Criar `infrastructure/temporal/workflows/infraction_reports/create_workflow.go`

  - Workflow: `CreateInfractionReportWorkflow`
  - Input struct: `CreateInfractionReportWorkflowInput`
  - Lógica:
    1. Execute CreateActivity
    2. Cache result
    3. Publish CoreEvents
    4. Publish DictEvents
    5. Start MonitorStatusWorkflow (child)

- [x] Criar `infrastructure/temporal/workflows/infraction_reports/update_workflow.go`

  - Workflow: `UpdateInfractionReportWorkflow`
  - Input struct: `UpdateInfractionReportWorkflowInput`
  - Lógica igual a CompleteClaimWorkflow

- [x] Criar `infrastructure/temporal/workflows/infraction_reports/cancel_workflow.go`

  - Workflow: `CancelInfractionReportWorkflow`
  - Input struct: `CancelInfractionReportWorkflowInput`
  - Lógica igual a CancelClaimWorkflow

- [x] Criar `infrastructure/temporal/workflows/infraction_reports/close_workflow.go`

  - Workflow: `CloseInfractionReportWorkflow`
  - Input struct: `CloseInfractionReportWorkflowInput`
  - Lógica igual a CompleteClaimWorkflow

- [x] Criar `infrastructure/temporal/workflows/infraction_reports/monitor_status_workflow.go`
  - Workflow: `MonitorInfractionReportStatusWorkflow`
  - Lógica similar a MonitorClaimStatusWorkflow
  - Monitorar status: CANCELLED, CLOSED
  - Usar GetInfractionReportActivity para polling

### 6.4 - Services

**Status**: ✅ COMPLETO

#### Tarefas:

- [x] Criar `infrastructure/temporal/services/infraction_report_service.go`
  - Struct `InfractionReportService`
  - Implementar interface `ports.InfractionReportService`
  - Métodos executam workflows via Temporal client

### 6.5 - UseCases

**Status**: ✅ COMPLETO

#### Tarefas:

- [x] Criar `application/usecases/infraction_report/application.go`

  - Struct `Application` com dependências
  - Constructor `NewApplication(service, obsProvider)`

- [x] Criar `application/usecases/infraction_report/create_infraction_report.go`

  - Método delega para service

- [x] Criar `application/usecases/infraction_report/update_infraction_report.go`

  - Método delega para service

- [x] Criar `application/usecases/infraction_report/cancel_infraction_report.go`

  - Método delega para service

- [x] Criar `application/usecases/infraction_report/close_infraction_report.go`
  - Método delega para service

### 6.6 - Handlers Pulsar

**Status**: ✅ COMPLETO

#### Tarefas:

- [x] Criar `handlers/pulsar/infraction_report/infraction_report_handler.go`

  - Struct `Handler` com `infractionReportApp`, `obsProvider`
  - Constructor `NewInfractionReportHandler`

- [x] Criar `handlers/pulsar/infraction_report/create_infraction_report_handler.go`

  - Método `CreateHandler`

- [x] Criar `handlers/pulsar/infraction_report/update_infraction_report_handler.go`

  - Método `UpdateHandler`

- [x] Criar `handlers/pulsar/infraction_report/cancel_infraction_report_handler.go`

  - Método `CancelHandler`

- [x] Criar `handlers/pulsar/infraction_report/close_infraction_report_handler.go`
  - Método `CloseHandler`

### 6.7 - Infrastructure gRPC

**Status**: ✅ COMPLETO

#### Tarefas:

- [x] Atualizar `infrastructure/grpc/gateway.go`

  - Adicionar `InfractionReportsClient`

- [x] Criar `infrastructure/grpc/infraction_report_client.go`
  - Inicializar client gRPC para InfractionReports

### 6.8 - Setup e Configuração

**Status**: ✅ COMPLETO

#### Tarefas:

- [x] Atualizar `setup/config.go`

  - Adicionar tópicos Pulsar:
    - `PulsarTopicCreateInfractionReport`
    - `PulsarTopicUpdateInfractionReport`
    - `PulsarTopicCancelInfractionReport`
    - `PulsarTopicCloseInfractionReport`

- [x] Atualizar `setup/pulsar.go`

  - Adicionar handler InfractionReport em `PulsarHandlers`
  - Subscrever nos novos tópicos
  - Adicionar rotas OnMessage

- [x] Atualizar `setup/temporal.go`

  - Registrar workflows de InfractionReport
  - Registrar activities de InfractionReport
  - Criar `InfractionReportService`

- [x] Atualizar `setup/setup.go`
  - Integrar InfractionReportHandler
  - Injetar dependências

### 6.9 - Validação Actions no SDK

**Status**: ✅ COMPLETO

#### Tarefas:

- [x] Verificar/adicionar constantes no SDK:
  - `pkg.ActionCreateInfractionReport` ✅
  - `pkg.ActionUpdateInformationInfractionReport` ✅
  - `pkg.ActionCancelInfractionReport` ✅
  - `pkg.ActionCloseInfractionReport` ✅

---

## Passo 7: Implementação orchestration-monitor

### 7.1 - Análise Completa

**Status**: ✅ COMPLETO (Passo 3)

### 7.2 - Activities

**Status**: ✅ COMPLETO

#### Arquivos Criados:

- [x] `infrastructure/temporal/activities/infraction_reports/infraction_report_activity.go`
- [x] `infrastructure/temporal/activities/infraction_reports/list_infraction_reports_activity.go`
- [x] `infrastructure/temporal/activities/infraction_reports/acknowledge_infraction_report_activity.go`

### 7.3 - Workflows

**Status**: ✅ COMPLETO

#### Arquivos Criados:

- [x] `infrastructure/temporal/workflows/infraction_reports/list_infraction_reports_workflow.go` (com polling e IsCounterparty=true)
- [x] `infrastructure/temporal/workflows/infraction_reports/acknowledge_infraction_report_workflow.go` (child workflow)

### 7.4 - Infrastructure gRPC

**Status**: ✅ COMPLETO

#### Tarefas:

- [x] Criar `infrastructure/grpc/infraction_report_client.go` (ListInfractionReports, AcknowledgeInfractionReport)
- [x] Atualizar `infrastructure/grpc/gateway.go` (InfractionReportsClient)

### 7.5 - Setup e Configuração

**Status**: ✅ COMPLETO

#### Tarefas:

- [x] Atualizar `setup/config.go` (CounterpartyParticipant, CursorKeyInfractionReport)
- [x] Atualizar `setup/temporal.go` (registros de 2 workflows + 2 activities)
- [x] Atualizar `setup/monitor_starter_process.go` (startInfractionReportsMonitor com IsCounterparty=true)

### 7.6 - Validação Build

**Status**: ✅ COMPLETO

```bash
cd apps/orchestration-monitor && go build -o /dev/null .  # ✅ SUCCESS
```

---

## Passo 8: Testes Unitários

### 8.1 - Testes Worker

**Status**: ⬜ PENDENTE

#### Tarefas:

- [ ] Criar `infrastructure/temporal/activities/infraction_reports/infraction_report_activity.go`

  - Struct `Activity` com `grpcGateway`
  - Constructor `NewActivity(grpcGateway)`

- [ ] Criar `infrastructure/temporal/activities/infraction_reports/list_infraction_reports_activity.go`

  - Activity: `ListInfractionReportsActivity`
  - Const: `ListInfractionReportsActivityName`

- [ ] Criar `infrastructure/temporal/activities/infraction_reports/acknowledge_infraction_report_activity.go`
  - Activity: `AcknowledgeInfractionReportActivity`
  - Const: `AcknowledgeInfractionReportActivityName`

### 7.3 - Workflows

**Status**: ⬜ PENDENTE

#### Tarefas:

- [ ] Criar `infrastructure/temporal/workflows/infraction_reports/acknowledge_infraction_report_workflow.go`

  - Workflow: `AcknowledgeInfractionReportWorkflow`
  - Input: `bacen.ExtendedInfractionReport`
  - Lógica:
    1. Execute AcknowledgeInfractionReportActivity
    2. Publish CoreEvents
    3. Publish DictEvents

- [ ] Criar `infrastructure/temporal/workflows/infraction_reports/list_infraction_reports_workflow.go`
  - Workflow: `ListInfractionReportsWorkflow`
  - Input struct: `ListInfractionReportsWorkflowInput`
  - Lógica similar a ListClaimsWorkflow:
    1. Buscar cursor (ModifiedAfter)
    2. Loop polling
    3. Para cada OPEN: ExecuteChildWorkflow(AcknowledgeInfractionReportWorkflow)
    4. Atualizar cursor
    5. ContinueAsNew

### 7.4 - Infrastructure gRPC

**Status**: ⬜ PENDENTE

#### Tarefas:

- [ ] Atualizar `infrastructure/grpc/gateway.go`

  - Adicionar `InfractionReportsClient`

- [ ] Criar `infrastructure/grpc/infraction_report_client.go` (se necessário)

### 7.5 - Setup e Configuração

**Status**: ⬜ PENDENTE

#### Tarefas:

- [ ] Atualizar `setup/config.go`

  - Adicionar:
    - `CounterpartyParticipant` (número do participante contraparte)
    - `CursorKeyInfractionReport` (chave para cursor)

- [ ] Atualizar `setup/temporal.go`

  - Registrar workflows InfractionReport
  - Registrar activities InfractionReport

- [ ] Atualizar `setup/monitor_starter_process.go`
  - Adicionar método para iniciar `ListInfractionReportsWorkflow`
  - Ou criar novo process separado

---

## Passo 8: Testes Unitários

### 8.1 - orchestration-worker

**Status**: ⬜ PENDENTE

#### Tarefas:

- [ ] Criar `tests/unit/infrastructure/temporal/activities/infraction_reports/helper_tests.go`
- [ ] Criar testes para cada activity:

  - [ ] `create_infraction_report_activity_test.go`
  - [ ] `update_infraction_report_activity_test.go`
  - [ ] `cancel_infraction_report_activity_test.go`
  - [ ] `close_infraction_report_activity_test.go`
  - [ ] `get_infraction_report_activity_test.go`

- [ ] Criar testes para cada workflow:
  - [ ] `create_infraction_report_workflow_test.go`
  - [ ] `update_infraction_report_workflow_test.go`
  - [ ] `cancel_infraction_report_workflow_test.go`
  - [ ] `close_infraction_report_workflow_test.go`
  - [ ] `monitor_infraction_report_status_workflow_test.go`

### 8.2 - orchestration-monitor

**Status**: ⬜ PENDENTE

#### Tarefas:

- [ ] Criar `tests/unit/temporal/activities/infraction_reports/helper_tests.go`
- [ ] Criar testes para activities:

  - [ ] `infraction_report_activity_test.go`
  - [ ] `list_infraction_reports_activity_test.go`
  - [ ] `acknowledge_infraction_report_activity_test.go`

- [ ] Criar testes para workflows:
  - [ ] `acknowledge_infraction_report_workflow_test.go`
  - [ ] `list_infraction_reports_workflow_test.go`

---

## Passo 9: Revisão e Validação Final

**Status**: ⬜ PENDENTE

### Tarefas:

- [ ] Revisar todas interfaces e injeções de dependências
- [ ] Validar nomenclatura e padrões
- [ ] Verificar mapeamentos gRPC ↔ Bacen
- [ ] Testar integração entre componentes
- [ ] Validar envs e configurações
- [ ] Executar linters (goimports, golangci-lint)
- [ ] Verificar cobertura de testes

---

## Notas de Implementação

### Diferenças importantes vs Claims:

1. **Update** usa `UpdateInformationInfractionReportRequest` (não é "complete")
2. **Close** é a operação de finalização (equivalente a "complete" em claims)
3. **Acknowledge** é o recebimento da notificação (equivalente a "acknowledge" em claims)
4. **Status**: OPEN → ACKNOWLEDGED → CLOSED/CANCELLED
5. **Monitor**: Lista com `IsCounterparty=true` para receber notificações (equivalente a `isDonor` em claims)

### Ordem sugerida de implementação (worker):

1. Ports e Activities básicas (create, get)
2. Workflow Create + MonitorStatus
3. Workflows Update, Cancel, Close
4. Handlers e UseCases
5. Setup e configuração

### Ordem sugerida de implementação (monitor):

1. Activities (list, close)
2. Workflows (close, list)
3. Setup e configuração
4. MonitorStarter integration

---

## Checklist Final

- [ ] Todas dependências injetadas corretamente
- [ ] Workflows registrados no Temporal worker
- [ ] Activities registradas com nomes corretos
- [ ] Handlers Pulsar configurados
- [ ] Tópicos Pulsar subscritos
- [ ] ENVs configuradas
- [ ] Testes unitários passando
- [ ] Linters executados sem erros
- [ ] Documentação atualizada
