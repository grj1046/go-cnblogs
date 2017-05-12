/*
update: 2015/05/10 17:27
table
    Ing
    Comment
    OriginIng
*/
CREATE TABLE IF NOT EXISTS `Ing` (
    `IngID` INTEGER PRIMARY KEY,
    `AuthorID` INTEGER,
    `AuthorUserName` VARCHAR(40),
    `AuthorNickName` VARCHAR(30),
    `Time` VARCHAR(25),
    `Status` VARCHAR(3),
    `Lucky` BOOL NOT NULL DEFAULT 0,
    `IsPrivate` BOOL NOT NULL DEFAULT 0,
    `AcquiredAt` VARCHAR(40),
    `Body` VARCHAR(300)
);

CREATE TABLE IF NOT EXISTS `Comment` (
    `ID` INTEGER PRIMARY KEY AUTOINCREMENT,
    `CommentID` INTEGER,
    `IngID` INTEGER,
	`AuthorID` INTEGER,
	`AuthorUserName` VARCHAR(40),
	`AuthorNickName` VARCHAR(30),
	`Body` VARCHAR(200),
	`Time` VARCHAR(25),
	`IsDelete` BOOL NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS `OriginIng` (
    `ID` INTEGER PRIMARY KEY AUTOINCREMENT,
    `IngID` INTEGER,
	`Status` VARCHAR(3),
	`AcquiredAt` VARCHAR(40),
	`Exception` TEXT,
    `HTMLHash`  VARCHAR(32),
	`HTML`       TEXT
);