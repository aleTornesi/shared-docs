-- name: GetDocuments :many
SELECT *
FROM documents d
LEFT JOIN document_access da ON d.id = da.document_id
LEFT JOIN page p ON d.id = p.document_id
WHERE da.user_id = $1 OR d.owner_id = $1
ORDER BY d.id DESC;

-- name: GetDocument :one
SELECT *
FROM documents d
LEFT JOIN document_access da ON d.id = da.document_id
WHERE d.id = $1 AND (da.user_id = $2 or d.owner_id = $2);

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

