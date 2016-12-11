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
