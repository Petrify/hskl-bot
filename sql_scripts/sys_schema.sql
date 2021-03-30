CREATE TABLE `service` (
  `serviceid` int NOT NULL,
  `servicename` varchar(20) NOT NULL,
  `servicetype` varchar(20) NOT NULL,
  `startupservice` tinyint NOT NULL DEFAULT '0',
  `version` int NOT NULL DEFAULT '0',
  PRIMARY KEY (`serviceid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;