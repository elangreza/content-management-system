ALTER TABLE
    article_versions
ADD
    COLUMN tag_relationship_score DOUBLE PRECISION NOT NULL DEFAULT 0.0;