-- Cakupan konteks disimpan pada percakapan: semua_catatan, satu_catatan, atau dokumen.
-- Percakapan lama memakai cakupan dokumen.
ALTER TABLE conversations ADD COLUMN cakupan TEXT NOT NULL DEFAULT 'dokumen'
    CHECK (cakupan IN ('semua_catatan', 'satu_catatan', 'dokumen'));
ALTER TABLE conversations ADD COLUMN catatan_id TEXT;
