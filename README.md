# Go RESTful API - Invoice resource

Olá, Stone! :tada:

Muito legal esse desafio. Aprendi demais. :sparkler:

Infelizmente, não tive tanto tempo quanto eu gostaria porque esse final de semestre está muito apertado (leia-se: dormir mais de 3 horas por dia é luxo).

A parte do servidor web eu praticamente fiz tudo ontem e hoje e mesmo assim acho que o resultado ficou bem legal! :smiley:

## API Syntax:

### GET /invoices
#### Parâmetros de query:
  - `apiToken`: token de autenticação
  - `document`: filtra os invoices pelo campo `Document`
    - validação de tamanho máximo: 14
    - validação de parâmetro duplicado
  - `referenceMonth`, `referenceYear`: filtra os invoices pelo capo ReferenceMonth e ReferenceYear
    - validação de tipo (verifica se é inteiro)
    - validação de parâmetro duplicado
  - `sort`: define a ordenação do resultado. Campos separados por vírgulas. Uso de `-` para indicar ordem decrescente
    * verificação da sintaxe
    * verificação dos campos selecionados (apenas `document`, `ReferenceMonth` e `ReferenceYear` são permitidos)
    * validação de parâmetro duplicado  
  - `page`, `perPage`: controlam a paginação
    * default: `page`=1, `perPage`=5
    * validação do número de página (1 <= `page` <= `lastPage`) (note que o valor de `lastPage` depende dos invoices na base, de `perPage` e dos filtros também)
    * Resposta com Header `X-Count-Total` para indicar quantidade total de invoices achados pela busca, independente de quanto são mostrados na página atual.
    * Resposta com Header `Link` indicando URI para próxima página, última página, primeira página e página anterior, quando aplicável.
    
#### Respostas:
  - `200, { "items": [listaDeInvoices] }`
  - `400, { "error": mensagemDeErro }`
  - `500`
    
#### Exemplos:
  - `localhost:3000/invoices?apiToken=sweetpotato`
  - `localhost:3000/invoices?document=JAkv92kLAFc&apiToken=sweetpotato`
  - `localhost:3000/invoices?sort=-referenceMonth&apiToken=sweetpotato`
  - `localhost:3000/invoices?document=JAkv92kLAFc&sort=-referenceYear,referenceMonth,-document&page=2&perPage=12&apiToken=sweetpotato`

### GET /invoices/:id

### POST /invoices

### PUT /invoices/:id

### DELETE /invoices/:id

## Pontos a destacar:

### Coisas legais:
- A estrutura do projeto. Busquei fazer de uma forma bem modular e organizada. Sendo meu primeiro projeto desse tipo dá pra imaginar como não é trivial conceber uma boa organização de primeira. Isso me tomou bastante tempo, mas gostei do resultado.
- Use de Headers como X-Total-Count, Link e Location nas respostas de buscas e inserções.
- Uso de arquivo de configuração .toml para guardar configurações, incluindo o token de aplicação.
- Uso de Middlewares para validação de dados e autenticação.
- Extensiva validação de dados para diversos dos casos: veja a validação feita para o Index de invoices, por exemplo.
- Leitura de mais de 500 páginas sobre Go e horas explorando a documentação.
- Uso de Buffer em algumas situações para acelerar a concatenação de strings.
- Comecei odiando o fato de ter que fazer *if err != nil {}* toda hora, mas depois aprendi a gostar.
- Uso de índices no banco de dados para otimizar.
- Analisei diversos frameworks web para Go antes de optar pelo Gin.
- Li várias discussões sobre quando usar o código 204, 400, 422... flame wars...

### Coisas tristes:
- Não concluí os testes. Criei toda estrutura necessária para criar o restante dos testes, mas vai me tomar muito mais tempo que não estou podendo dispor no momento. De qualquer forma, acredito que com os testes que criei já dá pra entender a forma como abordei o problema.
- Pro PUT eu só fiz atualizar a description. Fazer a validação pro resto tomaria bem mais tempo.

### Coisas extras (para o futuro):
- Versionar a api. Colocar url base '/v1/' para permitir novas versões da API no futuro.
- Verificar a possibilidade de melhor compressão com gzip.
- Implementar JWT para autenticação
- Implementar limite de requisições por usuário, usando cabeçalhos X-RateLimit-*
- Testar godep para gerenciar as dependências do projeto
- Usar reflection para fazer o Middleware do QueryOptions ser mais genérico
- Implementar o método PUT de forma adequada.
- Usar uma biblioteca para mock.
- Fazer migration do banco usando o próprio go.
- Documentar o código e a api.
- Estudar mais sobre RESTful para verificar como melhorar a API.
- Confirmar até que ponto a biblioteca padrão pra banco de dados implementa connection pooling.

## MySQL

A seguir, o código necessário para gerar o banco de dados usado pelo servidor.

```sql
CREATE DATABASE Stone COLLATE utf8_general_ci;

CREATE USER 'stone'@'localhost' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON Stone.* TO 'stone'@'localhost';
FLUSH PRIVILEGES;

USE Stone;

CREATE TABLE Invoice (
  Id INTEGER NOT NULL AUTO_INCREMENT,
  CreatedAt DATETIME NOT NULL,
  ReferenceMonth INTEGER NOT NULL,
  ReferenceYear INTEGER NOT NULL,
  Document VARCHAR(14) NOT NULL,
  Description VARCHAR(256) NOT NULL DEFAULT "",
  Amount DECIMAL(16, 2) NOT NULL,
  IsActive TINYINT NOT NULL DEFAULT 0,
  DeactiveAt DATETIME DEFAULT NULL,

  PRIMARY KEY (Id),
  /*? UNIQUE (Document) ?*/
  INDEX Document_Index (Document),
  INDEX ReferenceMonth_Index (ReferenceMonth),
  INDEX ReferenceYear_Index (ReferenceYear),
  INDEX IsActive_Index (IsActive)
);
```
[![baby-gopher](https://raw.githubusercontent.com/drnic/babygopher-site/gh-pages/images/babygopher-badge.png)](http://www.babygopher.org)
