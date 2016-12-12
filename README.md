# Go RESTful API - Invoice resource

Olá, Stone! :tada:

Muito legal esse desafio. Aprendi demais. :sparkler:

Infelizmente, não tive tanto tempo quanto eu gostaria porque esse final de semestre está muito apertado (leia-se: dormir mais de 3 horas por dia é luxo).

A parte do servidor eu praticamente fiz tudo ontem e hoje e mesmo assim acho que o resultado ficou bem legal! :smiley:

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

`localhost:3000/invoices/1?apiToken=sweetpotato`  
Response: `200` ou `404`  
```
{
  "item": {
    "id": 1,
    "createdAt": "2016-12-11T18:46:12Z",
    "referenceMonth": 12,
    "referenceYear": 2016,
    ..
  }
}
```

### POST /invoices

`localhost:3000/invoices?apiToken=sweetpotato`  
```
Body(form-data): {
  document: JdLCkji29SKl
  description: Lorem ipsum dolor sit amet.
  amount: 999.99
}
```  
Response: `201` Created  
Header `Location:` localhost:3000/invoices/42

### PUT /invoices/:id

`localhost:3000/invoices/1?description=abc&apiToken=sweetpotato`  
Response: `204` No Content

### DELETE /invoices/:id

`localhost:3000/invoices/1?apiToken=sweetpotato`  
Response: `204` ou `404`  

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

## Decidindo como implementar a API RESTful

Há vários frameworks em Go para auxiliar na criação do servidor web, alguns até mesmo auxiliando na criação de uma API RESTful.

- [sleepy](https://github.com/dougblack/sleepy): parado há 3 anos
- [gin](https://github.com/gin-gonic/gin): parece bem maduro, rico em funcionalidades
- [go-restful](https://github.com/emicklei/go-restful): desenvolvimento ativo, ótimos exemplos no repositório
- [web](https://github.com/hoisie/web): parado há 4 meses, exemplos no repositório são pobres
- [echo](https://github.com/labstack/echo): muito inicial, poucos exemplos
- [revel](http://revel.github.io/): parece bem completo, mas não preciso de tanto, parece overkill

Uma rápida pesquisa no Google nos retorna alguns tutoriais para criação de APIs RESTful:

- [gin](http://blog.narenarya.in/build-rest-api-go-mysql.html): recente (06/2016)
- [gorilla](http://www.giantflyingsaucer.com/blog/?p=5635): bom para estudar tratamento de erros
- ...

Decidi ir com [este](http://thenewstack.io/make-a-restful-json-api-go/) tutorial, pois me parece suficientemente bom para minhas necessidades e apenas faz uso de um Router externo (do (Gorilla toolkit](http://www.gorillatoolkit.org/)). O tutorial inclusive me parece adequado ao meu nível iniciante me Go.

Agradecimentos:
- http://www.restapitutorial.com/
- http://restful-api-design.readthedocs.io/

[![baby-gopher](https://raw.githubusercontent.com/drnic/babygopher-site/gh-pages/images/babygopher-badge.png)](http://www.babygopher.org)
