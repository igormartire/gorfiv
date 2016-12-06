# Go RESTful API - Invoice resource

## MySQL

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
  Description VARCHAR(256),
  Amount DECIMAL(16, 2) NOT NULL,
  IsActive TINYINT NOT NULL,
  DeactiveAt DATETIME,

  PRIMARY KEY (Id),
  /*? UNIQUE (Document) ?*/
  INDEX Document_Index (Document),
  INDEX ReferenceMonth_Index (ReferenceMonth),
  INDEX ReferenceYear_Index (ReferenceYear)
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
