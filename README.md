# Stress-Tester

Este projeto é uma ferramenta de teste de estresse para serviços HTTP. Ele permite que você envie um grande número de requisições para um endpoint específico, medindo o desempenho e a capacidade de resposta do serviço.

### Pré-requisitos

- [Go](https://golang.org/doc/install) 1.22 ou superior
- [Docker](https://docs.docker.com/get-docker/)

### Passos para Executar

1. **Clone o repositório:**

   ```sh
   git clone https://github.com/muriloabranches/stress-tester.git
   cd stress-tester
   ```

2. **Construa a imagem Docker:**

   ```sh
   docker build -t stress-tester .
   ```

3. **Execute o container Docker:**

   ```sh
   docker run --rm stress-tester -url <your_url> -requests 10 -concurrency 2 -method GET -timeout 30s
   ```

### Parâmetros de Linha de Comando

- `-url`: URL do serviço a ser testado.
- `-requests`: Número total de requisições.
- `-concurrency`: Número de requisições concorrentes.
- `-method`: Método HTTP a ser usado (GET, POST, etc.).
- `-body`: Corpo da requisição (para métodos como POST).
- `-header`: Cabeçalhos HTTP a serem incluídos na requisição (pode ser usado múltiplas vezes).
- `-timeout`: Timeout para cada requisição.

### Exemplo de Uso

```sh
docker run --rm stress-tester -url http://google.com -requests 10 -concurrency 2 -method GET -timeout 30s
```

Este comando executará um teste de estresse com 10 requisições, 2 concorrentes, usando o método GET e um timeout de 30 segundos por requisição.

### Exemplos de Teste

Um servidor de teste já está disponível na pasta `/test`. Para utilizá-lo, siga os passos abaixo:

1. **Execute o servidor de teste:**

   ```sh
   go run test/server.go
   ```

2. **Execute o stress-tester para diferentes endpoints:**

   - **Endpoint `/slow`:**

     ```sh
     docker run --rm stress-tester -url http://host.docker.internal:8080/slow -requests 10 -concurrency 2 -method GET -timeout 3s
     ```

   - **Endpoint `/success`:**

     ```sh
     docker run --rm stress-tester -url http://host.docker.internal:8080/success -requests 10 -concurrency 2 -method GET -timeout 30s
     ```

   - **Endpoint `/fail`:**

     ```sh
     docker run --rm stress-tester -url http://host.docker.internal:8080/fail -requests 10 -concurrency 2 -method GET -timeout 30s
     ```

   - **Endpoint `/with-body`:**

     ```sh
     docker run --rm stress-tester -url http://host.docker.internal:8080/with-body -requests 10 -concurrency 2 -method POST -body "test body" -timeout 30s
     ```

   - **Endpoint `/with-header`:**

     ```sh
     docker run --rm stress-tester -url http://host.docker.internal:8080/with-header -requests 10 -concurrency 2 -method GET -header "X-Test-Header: test" -timeout 30s
     ```

### Exemplo de Relatório Gerado

Após a execução do teste de estresse, um relatório será gerado e exibido no console. Além disso, um arquivo chamado `stress_test_report.txt` será criado no diretório atual com o conteúdo do relatório. Aqui está um exemplo de relatório gerado:

```
Stress-Tester Report
================================================
Execution Parameters:
  URL: http://localhost:8080/success
  HTTP Method: GET
  Body: 
  Headers:
    X-Test-Header: test
  Timeout: 30s
  Total number of requests desired: 10
  Concurrency level: 2
------------------------------------------------
Results:
  Total time taken: 50.105014s
  Total number of requests made: 10
  Number of requests with HTTP 200 status: 10
  Success rate: 100.00%
  Error rate: 0.00%
  Average time per request: 5.0105014s
  Distribution of other HTTP status codes:
================================================
```

Este relatório fornece uma visão geral dos parâmetros de execução e dos resultados do teste de estresse, incluindo o tempo total, a taxa de sucesso, a taxa de erro e o tempo médio por requisição.