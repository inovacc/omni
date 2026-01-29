o objetivo aqui é implementar os workflows para o contexto de refund.

veja esse app/orchestration-worker ele é uma aplicação que consome eventos de tópicos pulsar e executa workflows temporals. Dentro de workflows ele executa activities diversas, como chamar grpc, cachear, publicar eventos em topicos.

veja esse app/orchestration-monitor ele é uma aplicação que executa um monitoramento de status de diversos objetos. ao iniciar a aplicacao ele executa um ou mais workflows que ficam rodando e monitorando o status de objetos, quando o status muda, publica eventos em topicos pulsar.

1 - quero que voce analise o que ja foi implementado para /infraction_report, desde handlers, setups, applications, usecases, workflows, activities, mocks, envs, tests.

2 - depois eu quero que voce vá no repo de sdk-rsfn-validator e analise o libs/dict, la vai ter os contratos de refund, tanto grpc quanto structs para serem usadas em schemas etc.

4 - Quero que voce guarde guarde na pasta root .claude/context, em um markdown o que foi aprendido de relevante para possíveis outras tarefas futuras. e caso a janela de contexto acabe.

5 - Quero que voce crie um plano de ação a ser executado e guarde dentro de .claude/plans, para essa tarefa. e a cade execução marque lá o que foi feito.

6 - O que temos para implementar no worker são:

- Criar Solicitação de Devolução - parecido com create_infraction_report. seguir mesmo padrao. iremos criar a solicitacao de devolucao depois iremos iniciar um workflow para ficar consultando o status dessa solicitacao. esse workflow que ira monitorar seguir o mesmo padrao do MonitorInfractionReportStatusWorkflow. quando o status mudar para CANCELLED ou CLOSED, publicas eventos, assim como o MonitorInfractionReportStatusWorkflow faz.

- Consultar Solicitação de Devolução - ira criar a activity que sera usada nos workflows necessarios. seguir mesmo padrao da activity GetInfractionReportActivity.

- Cancelar Solicitação de Devolução - parecido com o CancelInfractionReportWorkflow. seguir mesmo padrao. executa activity de grpc dps publica os eventos.

- Fechar Solicitação de Devolução - parecido com o CloseInfractionReportWorkflow. seguir mesmo padrao. executa activity de grpc dps publica os eventos.

7 - Implementar no app/monitor

    7.1 - quero que voce analise o que ja foi implementado para /infraction_report, desde handlers, setups, applications, usecases, workflows, activities, mocks, envs, tests.
    7.2 - Quero que voce implemente o monitor de solicitacao de devolucao recebidas. esse monitor é parecido com o monitor de ListInfractionReportsWorkflow. ele vai listar as solicitacoes de devolucao criadas pro numero de participante que vai estar na env, e com ParticipantRole CONTESTED, status OPEN, e assim que tiver solicitacoes vai executar o workflow de Receber Solicitacao. Ou seja, implementar o workflow de Receber Solicitacao e chamar ele no monitor. siga o mesmo padrao do AcknowledgeInfractionReportWorkflow, porem sem a parte grpc, só os eventos.

8 - Depois de implementar tudo, quero que voce analise os testes unitarios ja criados na pasta /tests de cada aplicacao, orchestration-worker e orchestration-monitor, e crie os testes unitarios necessarios para cobrir as novas funcionalidades implementadas. analise apenas o de infraction_report para entender o padrao. e faça igual.

    8.1 - para os testes será necessario que voce analise o sdk-rsfn-validator/libs/dict para entender os contratos grpc e as structs que serao usadas nos testes. pois os testes precisam ter os requests completos para funcionar corretamente.

    8.2 - siga o mesmo padrao dos testes ja criados para infraction_report.

    8.3 - temos um make file em cada aplicacao, ali ele tera testes de cobertura. utilize se necessario.

    8.4 - preste atencao nos imports, muito das vezes os imports estao errados ou incompletos. por exemplo nao apontando para pasta correta do sdk-rsfn-validator/libs/dict.

    8.5 - preste atencao na cobertura, como o makefile fala, temos que estar acima de 80% de cobertura.

9 - Faça cada passo com cuidado, a cada final de cada passo me pergunte se quero que voce continue para o proximo passo. e a cada final de passo marque o proprio como finalizado no arquivo md de plano.

10 - Sempre que precisar pode retornar no arquivo de contexto para relembrar o que deve ser feito.

11 - quero que voce monitore a janela de contexto, e sempre que necessario guarde informações relevantes dentro de .claude/context. se ela estiver abaixo de 40% quero que voce crie um arquivo implementation*status*[date].md dentro de .claude/plan/status/[implemetation-name], e la voce vai documentar o que ja foi feito, o que falta fazer, e o que voce precisa para continuar.

Observação: Não esqueça de criar corretamente as interfaces e injecoes de dependencias. Tome cuidado para não duplicar o package na criação de um arquivo go.

Importante: think hard
