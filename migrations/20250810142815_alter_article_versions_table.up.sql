ALTER TABLE
    article_versions
ADD
    COLUMN relationship_score DOUBLE PRECISION NOT NULL DEFAULT 0.0;