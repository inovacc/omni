# Contexto de Implementação: Workflows de Refund

**Data:** 2025-10-31
**Objetivo:** Implementar workflows completos de Solicitação de Devolução (Refund) seguindo padrões de infraction_report

---

## 1. Resumo dos Contratos de Refund (SDK)

### 1.1 Operações gRPC Disponíveis

**Service:** `RefundService`

```go
- CreateRefund(CreateRefundRequest) → CreateRefundResponse
- GetRefund(GetRefundRequest) → GetRefundResponse
- ListRefunds(ListRefundsRequest) → ListRefundsResponse
- CloseRefund(CloseRefundRequest) → CloseRefundResponse
- CancelRefund(CancelRefundRequest) → CancelRefundResponse
```

### 1.2 Estruturas de Dados Principais

#### Refund (básica)
```go
type Refund struct {
    TransactionID string       // ID da transação PIX original
    RefundReason  RefundReason // Motivo da solicitação
    RefundAmount  float64      // Valor solicitado (>0)
    RefundDetails string       // Detalhes (max 2000 chars)
}
```

#### ExtendedRefund (completa)
```go
type ExtendedRefund struct {
    // Campos de Refund
    TransactionID string
    RefundReason  RefundReason
    RefundAmount  float64
    RefundDetails string

    // Campos estendidos
    ID                    string                 // ID único da solicitação
    Status                RefundStatus           // OPEN, CLOSED, CANCELLED
    ContestedParticipant  string                 // Participante contestado (8 dígitos)
    RequestingParticipant string                 // Participante solicitante (8 dígitos)
    CreationTime          *time.Time
    LastModified          *time.Time
    AnalysisResult        *RefundAnalysisResult  // Resultado da análise
    AnalysisDetails       *string
    RefundRejectionReason *RefundRejectionReason
    RefundTransactionID   *string                // ID da transação de devolução
}
```

### 1.3 Enums Importantes

#### RefundStatus
- `OPEN` - Solicitação aberta
- `CLOSED` - Solicitação fechada (analisada)
- `CANCELLED` - Solicitação cancelada

#### RefundReason
- `FRAUD` - Fraude
- `OPERATIONAL_FLAW` - Falha operacional
- `REFUND_CANCELLED` - Devolução cancelada
- `PIX_AUTOMATICO` - PIX Automático

#### RefundAnalysisResult
- `TOTALLY_ACCEPTED` - Totalmente aceito
- `PARTIALLY_ACCEPTED` - Parcialmente aceito
- `REJECTED` - Rejeitado

#### RefundRejectionReason
- `NO_BALANCE` - Sem saldo
- `ACCOUNT_CLOSURE` - Conta encerrada
- `INVALID_REQUEST` - Solicitação inválida
- `OTHER` - Outro motivo

#### RefundRequestRole
- `REQUESTING` - Solicitante (quem pede a devolução)
- `CONTESTED` - Contestado (quem deve devolver)

### 1.4 Requests e Responses

#### CreateRefundRequest
```go
type CreateRefundRequest struct {
    Signature   *string      // Assinatura digital (opcional)
    Participant string       // Participante solicitante (obrigatório)
    Refund      bacen.Refund // Dados da solicitação (obrigatório)
}
```

#### CloseRefundRequest
```go
type CloseRefundRequest struct {
    Signature              *string
    Participant            string                    // Obrigatório
    RefundID               string                    // Obrigatório
    RefundAnalysisResult   RefundAnalysisResult      // Obrigatório (TOTALLY_ACCEPTED, PARTIALLY_ACCEPTED, REJECTED)
    RefundAnalysisDetails  *string                   // Opcional
    RefundRejectionReason  *RefundRejectionReason    // Obrigatório se REJECTED
    RefundTransactionID    *string                   // Opcional (ID da transação de devolução)
}
```

#### CancelRefundRequest
```go
type CancelRefundRequest struct {
    Signature   *string
    Participant string  // Obrigatório
    RefundID    string  // Obrigatório
}
```

#### ListRefundsRequest
```go
type ListRefundsRequest struct {
    Participant                 string            // Obrigatório
    IncludeIndirectParticipants *bool             // Opcional
    ParticipantRole             RefundRequestRole // Obrigatório (REQUESTING, CONTESTED)
    Status                      []RefundStatus    // Opcional (filtro)
    IncludeDetails              *bool             // Opcional
    ModifiedAfter               *time.Time        // Opcional (para polling incremental)
    ModifiedBefore              *time.Time        // Opcional
    Limit                       *int              // Opcional
}
```

#### Response Pattern (todas as responses)
```go
type {Operation}Response struct {
    ResponseTime  time.Time        // Timestamp da resposta
    CorrelationID string           // ID de correlação
    Refund        ExtendedRefund   // Dados retornados (ou lista em ListRefunds)
}
```

### 1.5 Mappers Disponíveis

**Localização:** `github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/mappers/refund/`

- `create.go` - Mappers de CreateRefund
- `close.go` - Mappers de CloseRefund
- `cancel.go` - Mappers de CancelRefund
- `list.go` - Mappers de ListRefunds
- `get.go` - Mappers de GetRefund

**Pattern dos Mappers:**
```go
// Bacen → gRPC (para enviar request)
func MapBacen{Operation}RequestToGrpc(req *bacen.{Operation}Request) (*grpc.{Operation}Request, error)

// gRPC → Bacen (para processar response)
func MapGrpc{Operation}ResponseToBacen(resp *grpc.{Operation}Response) *bacen.{Operation}Response
```

---

## 2. Diferenças entre Refund e InfractionReport

### 2.1 Operações

| Operação | Refund | InfractionReport |
|----------|--------|------------------|
| Create | ✅ | ✅ |
| Get | ✅ | ✅ |
| List | ✅ | ✅ |
| Close | ✅ | ✅ |
| Cancel | ✅ | ✅ |
| Acknowledge | ❌ | ✅ |
| UpdateInformation | ❌ | ✅ |

**Refund NÃO possui:**
- `AcknowledgeRefund` - Não há reconhecimento de solicitação
- `UpdateInformation` - Não há atualização de informações após criação

### 2.2 Enums de Análise

**Refund:**
```go
RefundAnalysisResult:
- TOTALLY_ACCEPTED   (aceito totalmente)
- PARTIALLY_ACCEPTED (aceito parcialmente)
- REJECTED           (rejeitado)
```

**InfractionReport:**
```go
AnalysisResult:
- AGREED    (concorda com infração)
- DISAGREED (discorda da infração)
```

### 2.3 Campos Específicos de Refund

**Refund tem campos únicos:**
- `RefundAmount` (float64) - Valor da devolução solicitada
- `RefundTransactionID` - ID da transação de devolução (quando efetivada)
- `RefundRejectionReason` - Motivo da rejeição (se aplicável)

**InfractionReport tem:**
- `FraudType` - Tipo de fraude identificada
- Lógica de "acknowledge" antes de "close"

### 2.4 Fluxos de Trabalho

#### Refund
1. **Create** → Status: OPEN → **Monitor** inicia
2. Monitor detecta mudança de status → Publica eventos
3. **Close** (análise) ou **Cancel** → Status: CLOSED/CANCELLED

#### InfractionReport
1. **Create** → Status: OPEN → **Monitor** inicia
2. **Acknowledge** (reconhecimento)
3. **Update** (atualização de informações - opcional)
4. **Close** (análise) ou **Cancel** → Status: CLOSED/CANCELLED

**Conclusão:** Refund é mais simples, sem etapas intermediárias de acknowledge/update.

---

## 3. Padrões Arquiteturais Identificados

### 3.1 Estrutura de Arquivos (orchestration-worker)

```
application/
├── ports/{entity}.go                     # Interface do service
└── usecases/{entity}/
    ├── application.go                    # Struct com dependências
    ├── create_{entity}.go                # Use case de create
    ├── cancel_{entity}.go                # Use case de cancel
    └── close_{entity}.go                 # Use case de close

handlers/pulsar/{entity}/
├── {entity}_handler.go                   # Handler base
├── create_{entity}_handler.go            # Handler de create
├── cancel_{entity}_handler.go            # Handler de cancel
└── close_{entity}_handler.go             # Handler de close

infrastructure/
├── grpc/
│   ├── {entity}_client.go                # Cliente gRPC
│   └── gateway.go                        # Gateway centralizador
├── temporal/
│   ├── activities/{entities}/
│   │   ├── {entity}_activity.go          # Base activity
│   │   ├── create_activity.go            # Activity de create
│   │   ├── get_{entity}_activity.go      # Activity de get
│   │   ├── cancel_activity.go            # Activity de cancel
│   │   └── close_activity.go             # Activity de close
│   ├── workflows/{entities}/
│   │   ├── shared.go                     # Helpers compartilhados
│   │   ├── create_workflow.go            # Workflow de create
│   │   ├── monitor_status_workflow.go    # Workflow de monitoramento
│   │   ├── cancel_workflow.go            # Workflow de cancel
│   │   └── close_workflow.go             # Workflow de close
│   └── services/
│       └── {entity}_service.go           # Service que executa workflows

setup/
├── config.go                             # Configurações (env vars)
├── setup.go                              # Setup principal
├── temporal.go                           # Registro de workflows/activities
├── pulsar.go                             # Setup de consumer/producers
└── grpc.go                               # Setup do gRPC Gateway

tests/unit/infrastructure/temporal/
├── activities/{entities}/
│   ├── helper_tests.go                   # Mocks
│   ├── create_{entity}_activity_test.go
│   ├── get_{entity}_activity_test.go
│   ├── cancel_{entity}_activity_test.go
│   └── close_{entity}_activity_test.go
└── workflows/{entities}/
    ├── helper_test.go                    # Registro de stubs
    ├── create_{entity}_workflow_test.go
    ├── monitor_{entity}_status_workflow_test.go
    ├── cancel_{entity}_workflow_test.go
    └── close_{entity}_workflow_test.go
```

### 3.2 Fluxo de um Workflow Típico (Create)

```
1. Pulsar Message → Handler
2. Handler → Application Use Case
3. Application → Service (Temporal)
4. Service → ExecuteWorkflow (Temporal Client)
5. Workflow inicia:
   a. CreateActivity (gRPC) com GRPCOptions
      - Se falhar: NotifyFailure → CoreEvents
   b. CacheActivity com CacheOptions
      - Salva resposta no Redis
   c. CoreEventsPublishActivity com PublishEventOptions
      - Publica evento no tópico core-events
   d. DictEventsPublishActivity com PublishEventOptions
      - Publica evento no tópico dict-events
   e. MonitorStatusWorkflow (Child Workflow)
      - ParentClosePolicy: ABANDON
      - Loop de polling para detectar mudança de status
```

### 3.3 Monitor Status Workflow Pattern

```go
func MonitorRefundStatusWorkflow(ctx workflow.Context, createResp *Response) error {
    maxLoops := 1000
    checkInterval := 2 * time.Minute

    for i := 0; i < maxLoops; i++ {
        // Get current status
        resp, err := executeGetRefundActivity(ctx, participant, refundID)

        // Check if status changed
        if resp.Refund.Status == bacen.RefundStatusClosed ||
           resp.Refund.Status == bacen.RefundStatusCancelled {

            // Publish events
            ExecuteCoreEventsPublishActivity(ctx, refundID, ActionRefundStatusChanged, resp)
            ExecuteDictEventsPublishActivity(ctx, refundID, ActionRefundStatusChanged, resp)

            return nil // Workflow completo
        }

        workflow.Sleep(ctx, checkInterval)
    }

    // Continue as new to avoid history bloat
    return workflow.NewContinueAsNewError(ctx, MonitorRefundStatusWorkflow, createResp)
}
```

### 3.4 Tratamento de Erros em Activities

```go
func (a *Activity) CreateRefundActivity(ctx context.Context, req *refund.CreateRefundRequest) (*refund.CreateRefundResponse, error) {
    grpcResp, err := a.grpcGateway.RefundsClient.CreateRefund(ctx, req)
    if err != nil {
        // Classificar erro
        if utils.IsNonRetryableError(err) {
            // Erro de negócio (400, 404, 409, etc) - NÃO retentar
            return nil, temporal.NewNonRetryableApplicationError(
                "refund creation failed due to invalid request",
                "InvalidRequest",
                mappers.GrpcErrorToBacenProblem(err),
            )
        }
        // Erro transiente (500, timeout) - Permite retry
        return nil, mappers.GrpcErrorToBacenProblem(err)
    }

    return mapper.MapGrpcCreateRefundResponseToBacen(grpcResp), nil
}
```

**Políticas de Retry:**
- **GRPCOptions:** 5 tentativas, backoff 1.5x (3s → 30s)
- **CacheOptions:** 4 tentativas, backoff 1.5x (500ms → 5s)
- **PublishEventOptions:** 6 tentativas, backoff 1.5x (1s → 10s)

### 3.5 Estrutura de Testes

**Activity Test Pattern:**
```go
func TestCreateRefundActivity_Success(t *testing.T) {
    t.Parallel()

    ctx := context.Background()

    // Mock do gRPC client
    mockPB := &mockRefundServiceClient{
        createResp: &pb.CreateRefundResponse{...},
    }

    wrapper, _ := conngrpc.NewRefundGRPCClient(mockPB)
    gateway := &conngrpc.Gateway{RefundsClient: wrapper}
    activity := refunds.NewActivity(gateway)

    // Executar activity
    resp, err := activity.CreateRefundActivity(ctx, req)

    // Assertions
    assert.NoError(t, err)
    assert.NotNil(t, resp)
    assert.True(t, mockPB.calledCreate)
}
```

**Workflow Test Pattern:**
```go
func TestCreateRefundWorkflow_Success(t *testing.T) {
    t.Parallel()

    var suite testsuite.WorkflowTestSuite
    env := suite.NewTestWorkflowEnvironment()
    defer env.AssertExpectations(t)

    // Registrar stubs
    registerActivityStubsForRefunds(env)

    // Mock de activities específicas
    env.OnActivity(refundActivities.CreateRefundActivityName, mock.Anything, request).
        Return(&result, nil).Once()

    env.OnActivity(cacheActivities.CreateResponseCacheActivityName, mock.Anything, mock.Anything).
        Return(nil).Once()

    // ... outros mocks

    // Mock de child workflow
    env.OnWorkflow(wf.MonitorRefundStatusWorkflow, mock.Anything, &result).Return(nil)

    // Executar workflow
    env.ExecuteWorkflow(wf.CreateRefundWorkflow, input)

    // Assertions
    assert.NoError(t, env.GetWorkflowError())
}
```

---

## 4. Checklist de Componentes a Implementar

### 4.1 orchestration-worker

**Application Layer:**
- [ ] `application/ports/refund.go`
- [ ] `application/usecases/refund/application.go`
- [ ] `application/usecases/refund/create_refund.go`
- [ ] `application/usecases/refund/cancel_refund.go`
- [ ] `application/usecases/refund/close_refund.go`

**Handlers:**
- [ ] `handlers/pulsar/refund/refund_handler.go`
- [ ] `handlers/pulsar/refund/create_refund_handler.go`
- [ ] `handlers/pulsar/refund/cancel_refund_handler.go`
- [ ] `handlers/pulsar/refund/close_refund_handler.go`

**Infrastructure - gRPC:**
- [ ] `infrastructure/grpc/refund_client.go`
- [ ] `infrastructure/grpc/gateway.go` (adicionar RefundsClient)

**Infrastructure - Activities:**
- [ ] `infrastructure/temporal/activities/refunds/refund_activity.go`
- [ ] `infrastructure/temporal/activities/refunds/create_activity.go`
- [ ] `infrastructure/temporal/activities/refunds/get_refund_activity.go`
- [ ] `infrastructure/temporal/activities/refunds/cancel_activity.go`
- [ ] `infrastructure/temporal/activities/refunds/close_activity.go`

**Infrastructure - Workflows:**
- [ ] `infrastructure/temporal/workflows/refunds/shared.go`
- [ ] `infrastructure/temporal/workflows/refunds/create_workflow.go`
- [ ] `infrastructure/temporal/workflows/refunds/monitor_status_workflow.go`
- [ ] `infrastructure/temporal/workflows/refunds/cancel_workflow.go`
- [ ] `infrastructure/temporal/workflows/refunds/close_workflow.go`

**Infrastructure - Services:**
- [ ] `infrastructure/temporal/services/refund_service.go`

**Setup:**
- [ ] `setup/config.go` (adicionar vars)
- [ ] `setup/setup.go` (instanciar componentes)
- [ ] `setup/temporal.go` (registrar workflows/activities)
- [ ] `setup/pulsar.go` (adicionar topics e handlers)

**Tests - Activities:**
- [ ] `tests/unit/infrastructure/temporal/activities/refunds/helper_tests.go`
- [ ] `tests/unit/infrastructure/temporal/activities/refunds/create_refund_activity_test.go`
- [ ] `tests/unit/infrastructure/temporal/activities/refunds/get_refund_activity_test.go`
- [ ] `tests/unit/infrastructure/temporal/activities/refunds/cancel_refund_activity_test.go`
- [ ] `tests/unit/infrastructure/temporal/activities/refunds/close_refund_activity_test.go`

**Tests - Workflows:**
- [ ] `tests/unit/infrastructure/temporal/workflows/refunds/helper_test.go`
- [ ] `tests/unit/infrastructure/temporal/workflows/refunds/create_refund_workflow_test.go`
- [ ] `tests/unit/infrastructure/temporal/workflows/refunds/monitor_refund_status_workflow_test.go`
- [ ] `tests/unit/infrastructure/temporal/workflows/refunds/cancel_refund_workflow_test.go`
- [ ] `tests/unit/infrastructure/temporal/workflows/refunds/close_refund_workflow_test.go`

### 4.2 orchestration-monitor

**Infrastructure - gRPC:**
- [ ] `infrastructure/grpc/refund_client.go`
- [ ] `infrastructure/grpc/gateway.go` (adicionar RefundsClient)

**Infrastructure - Activities:**
- [ ] `infrastructure/temporal/activities/refunds/refund_activity.go`
- [ ] `infrastructure/temporal/activities/refunds/list_refunds_activity.go`

**Infrastructure - Workflows:**
- [ ] `infrastructure/temporal/workflows/refunds/list_refunds_workflow.go`
- [ ] `infrastructure/temporal/workflows/refunds/receive_refund_workflow.go`

**Setup:**
- [ ] `setup/config.go` (adicionar vars)
- [ ] `setup/temporal.go` (registrar workflows/activities)
- [ ] `setup/monitor_starter_process.go` (adicionar startRefundsMonitor)

**Tests:**
- [ ] `tests/unit/temporal/workflows/refunds/list_refunds_workflow_test.go`
- [ ] `tests/unit/temporal/workflows/refunds/receive_refund_workflow_test.go`

---

## 5. Actions e Eventos

**Constantes no SDK:** `github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/actions.go`

```go
const (
    ActionCreateRefund            Action = "create_refund"
    ActionGetRefund               Action = "get_refund"
    ActionListRefunds             Action = "list_refunds"
    ActionCloseRefund             Action = "close_refund"
    ActionCancelRefund            Action = "cancel_refund"
    ActionRefundStatusChanged     Action = "refund_status_changed"
    ActionReceiveRefund           Action = "receive_refund"  // Para monitor
)
```

**Uso nos Workflows:**
- Publicar eventos com action apropriada
- Usado em `MessageProperties.Action`
- Facilita rastreamento e correlação

---

## 6. Variáveis de Ambiente Necessárias

### orchestration-worker
```bash
# Pulsar Topics - Input (consumidos)
PULSAR_TOPIC_CREATE_REFUND="persistent://lb-conn/dict/orchestration-worker-create-refund"
PULSAR_TOPIC_CANCEL_REFUND="persistent://lb-conn/dict/orchestration-worker-cancel-refund"
PULSAR_TOPIC_CLOSE_REFUND="persistent://lb-conn/dict/orchestration-worker-close-refund"

# Pulsar Topics - Output (já existentes, reutilizados)
PULSAR_TOPIC_CORE_EVENTS="persistent://lb-conn/dict/core-events"
PULSAR_TOPIC_DICT_EVENTS="persistent://lb-conn/dict/dict-events"
```

### orchestration-monitor
```bash
# Monitor Config
CURSOR_KEY_REFUND="monitor:refund:last_modified"
COUNTERPARTY_PARTICIPANT="12345678"  # Participant ID do sistema
```

---

## 7. Pontos de Atenção

### 7.1 Imports
- **CUIDADO:** Imports devem apontar corretamente para o SDK
- Pacote correto: `github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/...`
- Verificar imports de:
  - `bacen/refund/` (requests/responses)
  - `grpc/refund/` (código gerado proto)
  - `mappers/refund/` (conversões)

### 7.2 Package Duplicado
- **NUNCA** declarar package duas vezes no mesmo arquivo
- Exemplo errado:
  ```go
  package refunds
  package refunds  // ERRO!
  ```

### 7.3 Injeção de Dependências
- Sempre usar construtores (NewXXX)
- Passar dependências via parâmetros
- Evitar variáveis globais

### 7.4 Cobertura de Testes
- Meta: **≥ 80%**
- Comando: `make ci-tests`
- Cobrir: success cases + error cases + edge cases

### 7.5 Activity Names
- Sempre definir constante: `const CreateRefundActivityName = "CreateRefundGRPCActivity"`
- Usar no registro: `activity.RegisterOptions{Name: CreateRefundActivityName}`
- Usar nos testes: `env.OnActivity(CreateRefundActivityName, ...)`

---

## 8. Referências Úteis

### Arquivos de Referência (Infraction Report)
- Activity: `apps/orchestration-worker/infrastructure/temporal/activities/infraction_reports/create_activity.go`
- Workflow: `apps/orchestration-worker/infrastructure/temporal/workflows/infraction_reports/create_workflow.go`
- Handler: `apps/orchestration-worker/handlers/pulsar/infraction_report/create_infraction_report_handler.go`
- Service: `apps/orchestration-worker/infrastructure/temporal/services/infraction_report_service.go`
- Test (Activity): `apps/orchestration-worker/tests/unit/infrastructure/temporal/activities/infraction_reports/create_infraction_report_activity_test.go`
- Test (Workflow): `apps/orchestration-worker/tests/unit/infrastructure/temporal/workflows/infraction_reports/create_infraction_report_workflow_test.go`

### SDK Paths
- Proto: `/Users/william/go/src/github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/proto/refund/`
- Bacen: `/Users/william/go/src/github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/bacen/refund/`
- Mappers: `/Users/william/go/src/github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/mappers/refund/`
- gRPC Generated: `/Users/william/go/src/github.com/lb-conn/sdk-rsfn-validator/libs/dict/pkg/grpc/refund/`

---

**Última Atualização:** 2025-10-31
**Status:** Documentação completa - Pronto para implementação
