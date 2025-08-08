ALTER TABLE
    articles
ADD
    COLUMN published_version_id INT REFERENCES article_versions("id"),
ADD
    COLUMN drafted_version_id INT REFERENCES article_versions("id"),
ADD
    COLUMN version_sequence INT DEFAULT 1;