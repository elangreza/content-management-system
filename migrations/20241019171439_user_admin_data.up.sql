INSERT INTO
    public .users (
        "id",
        "name",
        email,
        "password",
        "permission"
    )
VALUES
    (
        gen_random_uuid(),
        'admin movie 1',
        'admin1@cms.test',
        '$2a$10$72P7G4kPA8hSMoVXlzFahuJ68ME2AI6hFVx9GYpJ8a5skviniSmUG',
        3
    );