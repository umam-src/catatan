<script lang="ts">
  type Notebook = { id: string; title: string };
  type Source = {
    id: string;
    title: string;
    kind: string;
    content?: string;
    locator: string;
    checksum: string;
    metadata_json: string;
    created_at: string;
    updated_at: string;
  };

  let notebooks: Notebook[] = [];
  let sources: Source[] = [];
  let selectedNotebook = '';
  let activeSource: Source | null = null;
  let title = '';
  let file: File | null = null;
  let loading = true;
  let importing = false;
  let saving = false;
  let deleting = false;
  let error = '';
  let notice = '';

  async function api(path: string, init: RequestInit = {}) {
    const response = await fetch(path, init);
    if (!response.ok) throw new Error(await response.text() || 'Permintaan gagal.');
    return response;
  }

  async function load() {
    loading = true;
    error = '';
    try {
      const response = await api('/api/notebooks');
      notebooks = await response.json();
      if (!selectedNotebook && notebooks.length) selectedNotebook = notebooks[0].id;
      if (selectedNotebook) await loadSources();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Gagal memuat buku.';
    } finally {
      loading = false;
    }
  }

  async function loadSources() {
    if (!selectedNotebook) {
      sources = [];
      activeSource = null;
      return;
    }
    try {
      const response = await api('/api/notebooks/' + selectedNotebook + '/sources');
      sources = await response.json();
      activeSource = null;
      notice = '';
    } catch (err) {
      error = err instanceof Error ? err.message : 'Gagal memuat sumber.';
    }
  }

  async function pilihSumber(source: Source) {
    error = '';
    notice = '';
    try {
      const response = await api('/api/sources/' + source.id);
      activeSource = await response.json();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Gagal memuat sumber.';
    }
  }

  function pilihBerkas(event: Event) {
    const input = event.currentTarget as HTMLInputElement;
    file = input.files?.[0] ?? null;
    if (file && !title) title = file.name;
  }

  async function impor() {
    if (!selectedNotebook || !file || importing) return;
    importing = true;
    error = '';
    notice = '';
    try {
      const form = new FormData();
      form.set('file', file);
      if (title.trim()) form.set('title', title.trim());
      const response = await api('/api/notebooks/' + selectedNotebook + '/sources', { method: 'POST', body: form });
      const source: Source = await response.json();
      sources = [source, ...sources];
      activeSource = source;
      file = null;
      title = '';
      const input = document.getElementById('source-file') as HTMLInputElement | null;
      if (input) input.value = '';
      notice = 'Sumber berhasil diimpor.';
    } catch (err) {
      error = err instanceof Error ? err.message : 'Gagal mengimpor sumber.';
    } finally {
      importing = false;
    }
  }

  async function simpanJudul() {
    if (!activeSource || !title.trim() || saving) return;
    saving = true;
    error = '';
    notice = '';
    try {
      const response = await api('/api/sources/' + activeSource.id, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title: title.trim() })
      });
      const updated: Source = await response.json();
      activeSource = updated;
      sources = sources.map((source) => source.id === updated.id ? { ...source, title: updated.title, updated_at: updated.updated_at } : source);
      title = updated.title;
      notice = 'Judul sumber berhasil disimpan.';
    } catch (err) {
      error = err instanceof Error ? err.message : 'Gagal menyimpan judul sumber.';
    } finally {
      saving = false;
    }
  }

  async function hapusSumber() {
    if (!activeSource || deleting) return;
    if (!confirm('Hapus sumber ini? Isi sumber akan dihapus dari perangkat ini.')) return;
    deleting = true;
    error = '';
    notice = '';
    const deletedID = activeSource.id;
    try {
      await api('/api/sources/' + deletedID, { method: 'DELETE' });
      sources = sources.filter((source) => source.id !== deletedID);
      activeSource = null;
      title = '';
      notice = 'Sumber berhasil dihapus.';
    } catch (err) {
      error = err instanceof Error ? err.message : 'Gagal menghapus sumber.';
    } finally {
      deleting = false;
    }
  }

  function ukuran(source: Source) {
    try {
      const metadata = JSON.parse(source.metadata_json) as { size?: number };
      if (typeof metadata.size !== 'number') return 'Ukuran tidak diketahui';
      return new Intl.NumberFormat('id-ID').format(metadata.size) + ' byte';
    } catch {
      return 'Ukuran tidak diketahui';
    }
  }

  $: if (activeSource && !saving) title = activeSource.title;

  load();
</script>

<svelte:head><title>Sumber — Buku Catatan</title></svelte:head>

{#if loading}
  <main class="center"><p>Memuat…</p></main>
{:else}
  <main class="page">
    <header>
      <div>
        <div class="eyebrow">RUANG SUMBER</div>
        <h1>Sumber</h1>
        <p>Simpan bahan teks di perangkat ini untuk digunakan kembali tanpa jaringan.</p>
      </div>
      <a href="/">Kembali ke catatan</a>
    </header>

    {#if error}<div class="error" role="alert">{error}</div>{/if}
    {#if notice}<div class="notice" role="status">{notice}</div>{/if}

    <section class="toolbar">
      <label>Buku
        <select bind:value={selectedNotebook} onchange={loadSources}>
          {#each notebooks as notebook}<option value={notebook.id}>{notebook.title}</option>{/each}
        </select>
      </label>
      <label>Judul sumber
        <input bind:value={title} maxlength="200" placeholder="Nama sumber" />
      </label>
      <label>Berkas teks
        <input id="source-file" type="file" accept=".txt,.md,.csv,.json,text/plain,text/markdown,text/csv,application/json" onchange={pilihBerkas} />
      </label>
      <button class="primary" disabled={!file || !selectedNotebook || importing} onclick={impor}>{importing ? 'Mengimpor…' : 'Impor sumber'}</button>
    </section>

    <div class="content">
      <nav aria-label="Daftar sumber">
        <div class="heading">{sources.length} sumber</div>
        {#if !sources.length}
          <div class="empty">Belum ada sumber di buku ini.</div>
        {:else}
          {#each sources as source}
            <button class:chosen={activeSource?.id === source.id} class="source" onclick={() => pilihSumber(source)}>
              <strong>{source.title}</strong>
              <small>{ukuran(source)}</small>
            </button>
          {/each}
        {/if}
      </nav>

      <article>
        {#if activeSource}
          <div class="source-head">
            <div><div class="eyebrow">TEKS LOKAL</div><h2>{activeSource.title}</h2></div>
            <code title="SHA-256">{activeSource.checksum}</code>
          </div>
          <div class="metadata">
            <span>Jenis: {activeSource.kind}</span>
            <span>{ukuran(activeSource)}</span>
            <span>Diperbarui: {new Date(activeSource.updated_at).toLocaleString('id-ID')}</span>
          </div>
          <div class="actions">
            <label>Ubah judul
              <input bind:value={title} maxlength="200" aria-label="Judul sumber aktif" />
            </label>
            <button class="primary" disabled={!title.trim() || saving || deleting} onclick={simpanJudul}>{saving ? 'Menyimpan…' : 'Simpan judul'}</button>
            <button class="danger" disabled={saving || deleting} onclick={hapusSumber}>{deleting ? 'Menghapus…' : 'Hapus sumber'}</button>
          </div>
          <pre>{activeSource.content}</pre>
        {:else}
          <div class="empty detail"><h2>Pilih sumber</h2><p>Daftar di kiri menampilkan bahan yang sudah disimpan di buku ini.</p></div>
        {/if}
      </article>
    </div>
  </main>
{/if}

<style>
  :global(*) { box-sizing: border-box; }
  :global(body) { margin: 0; font-family: Inter, ui-sans-serif, system-ui, sans-serif; background: #f7f8fa; color: #202124; }
  button, input, select { font: inherit; }
  button:focus-visible, input:focus-visible, select:focus-visible { outline: 2px solid #202124; outline-offset: 2px; }
  .center { min-height: 100vh; display: grid; place-items: center; }
  .page { min-height: 100vh; padding: 40px; max-width: 1320px; margin: auto; }
  header { display: flex; justify-content: space-between; gap: 24px; align-items: start; margin-bottom: 24px; }
  h1, h2 { margin: 6px 0; }
  header p { margin: 0; color: #686c72; }
  header a { color: #202124; }
  .eyebrow { font-size: 11px; letter-spacing: .12em; color: #73777c; font-weight: 700; }
  .error, .notice { margin: 0 0 16px; padding: 10px 12px; border-radius: 8px; }
  .error { background: #fff0ef; color: #8a1c13; }
  .notice { background: #eef7ee; color: #245b2a; }
  .toolbar { display: grid; grid-template-columns: 1fr 1fr 1.5fr auto; gap: 12px; align-items: end; padding: 16px; background: #fff; border: 1px solid #e2e5e8; border-radius: 12px; margin-bottom: 20px; }
  label { display: grid; gap: 6px; font-size: 13px; font-weight: 600; }
  input, select { width: 100%; border: 1px solid #d5d9dd; border-radius: 8px; padding: 9px 10px; background: #fff; }
  .primary, .danger { border: 0; border-radius: 8px; padding: 10px 14px; cursor: pointer; }
  .primary { background: #202124; color: #fff; }
  .danger { background: #fff0ef; color: #8a1c13; border: 1px solid #e4b5b0; }
  .primary:disabled, .danger:disabled { opacity: .45; cursor: not-allowed; }
  .content { min-height: 620px; display: grid; grid-template-columns: 300px 1fr; background: #fff; border: 1px solid #e2e5e8; border-radius: 12px; overflow: hidden; }
  nav { border-right: 1px solid #e2e5e8; padding: 12px; }
  .heading { padding: 8px 10px 12px; font-size: 13px; color: #686c72; }
  .source { width: 100%; display: grid; gap: 5px; text-align: left; border: 0; border-radius: 8px; background: transparent; padding: 11px 10px; cursor: pointer; }
  .source:hover, .source.chosen { background: #f1f3f4; }
  .source small { color: #73777c; }
  .empty { padding: 24px 10px; color: #73777c; }
  article { padding: 28px; min-width: 0; }
  .source-head { display: flex; justify-content: space-between; gap: 20px; align-items: start; }
  code { max-width: 420px; overflow-wrap: anywhere; color: #686c72; font-size: 12px; }
  .metadata { display: flex; gap: 18px; flex-wrap: wrap; color: #73777c; font-size: 13px; margin: 12px 0 20px; }
  .actions { display: grid; grid-template-columns: minmax(220px, 1fr) auto auto; gap: 10px; align-items: end; padding: 14px; margin-bottom: 20px; background: #f7f8fa; border: 1px solid #e2e5e8; border-radius: 10px; }
  pre { white-space: pre-wrap; overflow-wrap: anywhere; line-height: 1.65; margin: 0; font: inherit; }
  .detail { min-height: 500px; display: grid; place-content: center; text-align: center; }
  @media (max-width: 900px) {
    .page { padding: 20px; }
    .toolbar, .actions { grid-template-columns: 1fr; }
    .content { grid-template-columns: 1fr; }
    nav { border-right: 0; border-bottom: 1px solid #e2e5e8; max-height: 300px; overflow: auto; }
    header { flex-direction: column; }
    .source-head { flex-direction: column; }
  }
</style>
