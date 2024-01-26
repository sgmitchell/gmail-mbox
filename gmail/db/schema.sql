-- general info about messages
CREATE TABLE IF NOT EXISTS messages
(
    MessageID TEXT NOT NULL PRIMARY KEY,
    ThreadID  TEXT,
    FromName    TEXT,
    FromEmail TEXT,
    ToLine      TEXT,
    Subject   TEXT,
    SizeBytes INTEGER,
    Parts     INTEGER

);

-- map labels to messages
CREATE TABLE IF NOT EXISTS labels
(
    MessageID TEXT NOT NULL,
    Label     TEXT NOT NULL,
    PRIMARY KEY (MessageID, Label)
);

-- bodies is the plain and html content of an email
CREATE TABLE IF NOT EXISTS bodies
(
    MessageID TEXT NOT NULL PRIMARY KEY,
    PlainText TEXT,
    HTML      TEXT
);

-- all the parts in a message
CREATE TABLE IF NOT EXISTS parts
(
    MessageID TEXT    NOT NULL,
    PartIdx   INTEGER NOT NULL,
    Headers   TEXT,
    ContentSize INTEGER,
    --Content   BLOB,
    PRIMARY KEY (MessageID, PartIdx)
);
