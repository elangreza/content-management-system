INSERT INTO
    users (
        "id",
        "name",
        email,
        "password",
        "role"
    )
VALUES
    (
        '01988791-a211-7d89-9e13-4a185d429b05',
        'content writer',
        'contentwriter@cms.test',
        '$2a$10$EDMYqMCwEvn92qYKCczkr.68Q/pkegypFHzD4vLv6io37JjrmS4bi',
        3
    ),
    (
        '01988791-a211-7d89-9e13-4a185d429000',
        'editor',
        'editor@cms.test',
        '$2a$10$EDMYqMCwEvn92qYKCczkr.68Q/pkegypFHzD4vLv6io37JjrmS4bi',
        15
    );

INSERT INTO
    articles (
        id,
        created_by,
        created_at,
        updated_by,
        updated_at,
        version_sequence
    )
VALUES
    (
        1,
        '01988791-a211-7d89-9e13-4a185d429b05' :: uuid,
        '2025-08-11 11:42:40.710',
        '01988791-a211-7d89-9e13-4a185d429b05' :: uuid,
        '2025-08-11 12:04:16.108',
        3
    );

INSERT INTO
    article_versions (
        id,
        article_id,
        title,
        body,
        "version",
        status,
        created_by,
        created_at,
        updated_by,
        updated_at,
        tag_relationship_score
    )
VALUES
    (
        2,
        1,
        'string',
        'string',
        2,
        2,
        '01988791-a211-7d89-9e13-4a185d429b05' :: uuid,
        '2025-08-11 12:01:57.718',
        '01988791-a211-7d89-9e13-4a185d429000' :: uuid,
        '2025-08-11 12:03:18.961',
        1.0
    ),
    (
        1,
        1,
        'string',
        'string',
        1,
        1,
        '01988791-a211-7d89-9e13-4a185d429b05' :: uuid,
        '2025-08-11 11:42:40.710',
        '01988791-a211-7d89-9e13-4a185d429000' :: uuid,
        '2025-08-11 12:04:07.820',
        0.0
    ),
    (
        3,
        1,
        'string',
        'string 2',
        3,
        0,
        '01988791-a211-7d89-9e13-4a185d429b05' :: uuid,
        '2025-08-11 12:04:16.108',
        NULL,
        '2025-08-11 12:04:16.118',
        0.0
    );

UPDATE
    articles
SET
    created_by = '01988791-a211-7d89-9e13-4a185d429b05' :: uuid,
    created_at = '2025-08-11 11:42:40.710',
    updated_by = '01988791-a211-7d89-9e13-4a185d429b05' :: uuid,
    updated_at = '2025-08-11 12:04:16.108',
    published_version_id = 1,
    drafted_version_id = 3,
    archived_version_id = 2,
    version_sequence = 3
WHERE
    id = 1;

INSERT INTO
    tags ("name", created_at, updated_at)
VALUES
    ('string', '2025-08-11 11:42:40.710', NULL),
    ('string2', '2025-08-11 12:01:57.718', NULL);

INSERT INTO
    article_version_tags (tag_name, article_version_id, created_at)
VALUES
    ('string', 1, '2025-08-11 11:42:40.710'),
    ('string', 2, '2025-08-11 12:01:57.718'),
    ('string2', 2, '2025-08-11 12:01:57.718'),
    ('string', 3, '2025-08-11 12:04:16.108'),
    ('string2', 3, '2025-08-11 12:04:16.108');