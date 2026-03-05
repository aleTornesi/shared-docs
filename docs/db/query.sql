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

-- name: GetPage :one
SELECT *
FROM page p
WHERE p.document_id = $1 AND p.page_number = $2;


-- name: DeletePage :exec
DELETE FROM page
WHERE document_id = $1 AND page_number = $2;

-- name: DeleteDocument :exec
DELETE FROM documents
WHERE id = $1;

