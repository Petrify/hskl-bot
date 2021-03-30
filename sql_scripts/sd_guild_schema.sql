CREATE TABLE `user` (
  `iduser` varchar(20) NOT NULL,
  PRIMARY KEY (`iduser`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
CREATE TABLE `role` (
  `idrole` varchar(20) NOT NULL,
  PRIMARY KEY (`idrole`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
CREATE TABLE `channel` (
  `idchannel` varchar(20) NOT NULL,
  `fk_idrole` varchar(20) NOT NULL,
  PRIMARY KEY (`idchannel`),
  KEY `role_idx` (`fk_idrole`),
  CONSTRAINT `chan_role` FOREIGN KEY (`fk_idrole`) REFERENCES `role` (`idrole`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
CREATE TABLE `final` (
  `idfinal` int NOT NULL,
  `type` varchar(3) NOT NULL,
  `date` date DEFAULT NULL,
  `fk_idchannel` varchar(20) DEFAULT NULL,
  PRIMARY KEY (`idfinal`),
  KEY `final_channel_idx` (`fk_idchannel`),
  CONSTRAINT `final_channel` FOREIGN KEY (`fk_idchannel`) REFERENCES `channel` (`idchannel`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
CREATE TABLE `major` (
  `abbreviation` varchar(10) NOT NULL,
  `name` varchar(45) NOT NULL,
  PRIMARY KEY (`abbreviation`),
  UNIQUE KEY `name_UNIQUE` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
CREATE TABLE `module` (
  `abbr` varchar(10) NOT NULL,
  `fk_major` varchar(10) NOT NULL,
  `name` varchar(128) NOT NULL,
  `idmodule` int NOT NULL AUTO_INCREMENT,
  `semester` int NOT NULL,
  `fk_idfinal` int NOT NULL,
  PRIMARY KEY (`abbr`,`fk_major`),
  UNIQUE KEY `idmodule_UNIQUE` (`idmodule`),
  KEY `final_idx` (`fk_idfinal`),
  KEY `major_idx` (`fk_major`),
  KEY `idmodule_idx` (`idmodule`),
  CONSTRAINT `final` FOREIGN KEY (`fk_idfinal`) REFERENCES `final` (`idfinal`) ON DELETE CASCADE,
  CONSTRAINT `major` FOREIGN KEY (`fk_major`) REFERENCES `major` (`abbreviation`) ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=563 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
CREATE TABLE `ref_user_has_final` (
  `iduser` varchar(20) NOT NULL,
  `idfinal` int NOT NULL,
  PRIMARY KEY (`iduser`,`idfinal`),
  KEY `final_idx` (`idfinal`),
  CONSTRAINT `fk_final` FOREIGN KEY (`idfinal`) REFERENCES `final` (`idfinal`),
  CONSTRAINT `fk_user` FOREIGN KEY (`iduser`) REFERENCES `user` (`iduser`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
CREATE TABLE `option` (
  `key` varchar(128) NOT NULL,
  `value` varchar(128) DEFAULT NULL,
  `set_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `set_by` varchar(128) DEFAULT NULL,
  PRIMARY KEY (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
INSERT INTO `option` (`key`, `value`) VALUES ('command_prefix', '!');
INSERT INTO `option` (`key`, `value`) VALUES ('finalsCategoryID', '');