-- name: GetDocuments :many
SELECT d.id, d.title, d.owner_id, u.username, pages.c
FROM documents d
INNER JOIN users u ON d.owner_id = u.id
LEFT JOIN LATERAL (
    SELECT count(*) AS c FROM page WHERE document_id = d.id
) pages ON true
WHERE d.owner_id = $1 OR EXISTS (
    SELECT 1
    FROM document_access da
    WHERE da.user_id = $1 AND da.document_id = d.id
)
ORDER BY d.id DESC;

-- name: GetDocument :many
SELECT d.id, d.title, d.owner_id, u.username, p.page_number, p.content
FROM documents d
INNER JOIN users u ON d.owner_id = u.id
LEFT JOIN page p ON d.id = p.document_id
WHERE d.id = $1 and (
    d.owner_id = $2 OR EXISTS (
        SELECT 1
        FROM document_access da
        WHERE da.user_id = $2 AND da.document_id = d.id
    )
)
ORDER BY p.page_number ASC;

-- name: CreateDocument :one
INSERT INTO documents (title, owner_id)
VALUES ($1, $2)
RETURNING id;

-- name: PutTitle :exec
UPDATE documents
SET title = $1
WHERE id = $2 AND (
    owner_id = $3 OR
    EXISTS (
        SELECT 1
        FROM document_access
        WHERE user_id = $3 AND document_id = $2
    )
);

-- name: ValidateAccess :one
SELECT true
FROM documents d
WHERE d.id = $2 AND (
    d.owner_id = $1 OR EXISTS (
        SELECT 1
        FROM document_access
        WHERE document_id = $2 AND user_id = $1
    )
);

-- name: UpdatePageNumbers :exec
UPDATE page
SET page_number = page_number + 1
WHERE document_id = $1 AND page_number >= $2;

-- name: CreatePage :exec
INSERT INTO page (document_id, page_number, content)
    VALUES ($1, $2, '');
