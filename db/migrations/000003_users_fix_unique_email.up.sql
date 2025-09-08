-- 1) Bersihkan constraint lama YANG MUNGKIN tidak ada (aman)
ALTER TABLE users DROP CONSTRAINT IF EXISTS uni_users_email;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_idx;  -- kalau pernah salah nama

-- 2) Jika ada duplikat, unique akan gagal. Opsional: bersihkan dulu duplikat (simpan satu baris/ email).
--    Comment baris di bawah jika tidak ingin auto-delete.
WITH dups AS (
  SELECT ctid
  FROM (
    SELECT ctid,
           ROW_NUMBER() OVER (PARTITION BY LOWER(email) ORDER BY created_at NULLS LAST, id) AS rn
    FROM users
  ) t
  WHERE t.rn > 1
)
DELETE FROM users u
USING dups d
WHERE u.ctid = d.ctid;

-- 3) Tambahkan constraint unik yang benar bila belum ada
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint c
    WHERE c.conrelid = 'public.users'::regclass
      AND c.contype = 'u'
      AND c.conname  = 'users_email_key'
  ) THEN
    ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);
  END IF;
END$$;
