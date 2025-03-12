-- --------------------------------------------------------
-- Host:                         127.0.0.1
-- Server version:               9.2.0 - MySQL Community Server - GPL
-- Server OS:                    Win64
-- HeidiSQL Version:             12.10.0.7000
-- --------------------------------------------------------

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET NAMES utf8 */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;


-- Dumping database structure for movies
CREATE DATABASE IF NOT EXISTS `movies` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci */ /*!80016 DEFAULT ENCRYPTION='N' */;
USE `movies`;

-- Dumping structure for table movies.comments
CREATE TABLE IF NOT EXISTS `comments` (
  `commentId` int unsigned NOT NULL AUTO_INCREMENT,
  `movieId` int unsigned NOT NULL DEFAULT '0',
  `userId` int unsigned DEFAULT NULL,
  `comment` varchar(1000) NOT NULL DEFAULT '',
  `postedDT` datetime NOT NULL DEFAULT (now()),
  PRIMARY KEY (`commentId`),
  UNIQUE KEY `commentId` (`commentId`),
  KEY `FK_comments_movies` (`movieId`),
  KEY `userId` (`userId`),
  CONSTRAINT `FK_comments_movies` FOREIGN KEY (`movieId`) REFERENCES `movies` (`movieId`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `FK_comments_users` FOREIGN KEY (`userId`) REFERENCES `users` (`userId`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=125 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Data exporting was unselected.

-- Dumping structure for procedure movies.DeleteComment
DELIMITER //
CREATE PROCEDURE `DeleteComment`(
	IN `userId` INT,
	IN `commentId` INT
)
BEGIN
	WITH a AS (
		SELECT
			userId,
			admin		 
		FROM users
		WHERE users.userId = userId
	)
	DELETE FROM comments c
	WHERE c.commentId = commentId AND (c.userId = userId OR (SELECT admin FROM a) = 1);
END//
DELIMITER ;

-- Dumping structure for procedure movies.GetComments
DELIMITER //
CREATE PROCEDURE `GetComments`(
	IN `movieId` INT,
	IN `lastCommentId` INT,
	IN `n` INT
)
BEGIN
	WITH c AS (
		SELECT *
		FROM comments
		WHERE comments.movieId = movieId AND IF(lastCommentID = 0, TRUE, comments.commentId < lastCommentId)
		ORDER BY commentId DESC
		LIMIT n
	)
	SELECT c.comment,
		 	 c.postedDT,
		 	 c.userId,
		 	 c.commentId,
		 	 IFNULL(users.username, 'DELETED')
	FROM c
	LEFT JOIN users ON c.userId = users.userId;
END//
DELIMITER ;

-- Dumping structure for procedure movies.GetLatestComments
DELIMITER //
CREATE PROCEDURE `GetLatestComments`(
	IN `n` INT
)
BEGIN
	SELECT 
		c.commentId,
		c.comment,
		u.username,
		u.userId,
		m.movieId,
		m.title
	FROM comments c
	JOIN movies m ON c.movieId = m.movieId
	JOIN users u ON c.userId = u.userId
	ORDER BY commentId DESC
	LIMIT n;
END//
DELIMITER ;

-- Dumping structure for procedure movies.GetLatestMovies
DELIMITER //
CREATE PROCEDURE `GetLatestMovies`(
	IN `n` INT
)
BEGIN
	SELECT 
		m.movieId, 
		m.title, 
		IFNULL(AVG(r.rating), 0) 
	FROM movies m 
	LEFT JOIN movierating r ON m.movieId = r.movieId 
	GROUP BY m.movieId
	ORDER BY m.movieId DESC
	LIMIT n;
END//
DELIMITER ;

-- Dumping structure for procedure movies.GetMovie
DELIMITER //
CREATE PROCEDURE `GetMovie`(
	IN `movieId` INT
)
BEGIN
	SELECT
		m.movieId,
		m.title,
		m.genres,
		IFNULL(AVG(r.rating), 0) AS avgRating,
		COUNT(r.rating) AS nRatings
	FROM movies m
	LEFT JOIN movierating r ON r.movieId = m.movieId
	WHERE m.movieId = movieId
	GROUP BY m.movieId;
END//
DELIMITER ;

-- Dumping structure for table movies.movierating
CREATE TABLE IF NOT EXISTS `movierating` (
  `userId` int unsigned NOT NULL,
  `movieId` int unsigned NOT NULL,
  `rating` decimal(2,1) unsigned NOT NULL DEFAULT '0.0',
  `timeStamp` datetime NOT NULL DEFAULT (now()),
  UNIQUE KEY `userId_movieId` (`userId`,`movieId`) USING BTREE,
  KEY `movieId_rating` (`movieId`,`rating`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Data exporting was unselected.

-- Dumping structure for table movies.movies
CREATE TABLE IF NOT EXISTS `movies` (
  `movieId` int unsigned NOT NULL AUTO_INCREMENT,
  `title` varchar(255) NOT NULL,
  `genres` varchar(255) DEFAULT NULL,
  `addedDT` datetime DEFAULT (now()),
  `adderUserId` int unsigned DEFAULT NULL,
  PRIMARY KEY (`movieId`),
  UNIQUE KEY `movieId_UNIQUE` (`movieId`),
  KEY `userId` (`adderUserId`),
  CONSTRAINT `userId` FOREIGN KEY (`adderUserId`) REFERENCES `users` (`userId`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=209180 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Data exporting was unselected.

-- Dumping structure for table movies.sessions
CREATE TABLE IF NOT EXISTS `sessions` (
  `token` varchar(64) NOT NULL,
  `expirationDT` datetime NOT NULL,
  `userId` int unsigned NOT NULL DEFAULT (0),
  PRIMARY KEY (`token`),
  UNIQUE KEY `token` (`token`),
  KEY `userId` (`userId`),
  CONSTRAINT `userIdFK` FOREIGN KEY (`userId`) REFERENCES `users` (`userId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Data exporting was unselected.

-- Dumping structure for procedure movies.SetComment
DELIMITER //
CREATE PROCEDURE `SetComment`(
	IN `userId` INT,
	IN `commentId` INT,
	IN `comment` VARCHAR(1000)
)
BEGIN
	WITH a AS (
		SELECT
			userId,
			admin		 
		FROM users
		WHERE users.userId = userId
	)
	UPDATE comments c
	SET c.comment = comment
	WHERE c.commentId = commentId AND (c.userId = userId OR (SELECT admin FROM a) = 1);
END//
DELIMITER ;

-- Dumping structure for table movies.users
CREATE TABLE IF NOT EXISTS `users` (
  `userId` int unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(45) NOT NULL,
  `password` varchar(255) NOT NULL,
  `registerDate` datetime NOT NULL DEFAULT (now()),
  `admin` tinyint(1) NOT NULL DEFAULT (0),
  `banUntil` datetime NOT NULL DEFAULT (now()),
  PRIMARY KEY (`userId`,`username`),
  UNIQUE KEY `userId_UNIQUE` (`userId`),
  UNIQUE KEY `userscol_UNIQUE` (`password`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Data exporting was unselected.

/*!40103 SET TIME_ZONE=IFNULL(@OLD_TIME_ZONE, 'system') */;
/*!40101 SET SQL_MODE=IFNULL(@OLD_SQL_MODE, '') */;
/*!40014 SET FOREIGN_KEY_CHECKS=IFNULL(@OLD_FOREIGN_KEY_CHECKS, 1) */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40111 SET SQL_NOTES=IFNULL(@OLD_SQL_NOTES, 1) */;
