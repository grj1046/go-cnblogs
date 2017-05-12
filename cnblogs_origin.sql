
CREATE TABLE IF NOT EXISTS `OriginIng` (
    `ID` INTEGER PRIMARY KEY AUTOINCREMENT,
    `IngID` INTEGER,
	`Status` VARCHAR(3),
	`AcquiredAt` VARCHAR(40),
	`Exception` TEXT,
    `HTMLHash`  VARCHAR(32),
	`HTML`       TEXT
);