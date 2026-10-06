<script lang="ts">
  import { onMount } from 'svelte';

  let { children } = $props();
  type User = { username: string; display_name: string };

  let user: User | null = null;
  let menuPenggunaTerbuka = false;
  let dialogPengaturanTerbuka = false;
  let pendaftaranDiizinkan = true;

  const kunciPendaftaran = 'catatan.pendaftaranDiizinkan';

  function muatPengaturan() {
    const nilai = localStorage.getItem(kunciPendaftaran);
    pendaftaranDiizinkan = nilai === null ? true : nilai === 'true';
    document.documentElement.toggleAttribute('data-pendaftaran-ditutup', !pendaftaranDiizinkan);
  }

  function simpanPengaturan() {
    localStorage.setItem(kunciPendaftaran, String(pendaftaranDiizinkan));
    document.documentElement.toggleAttribute('data-pendaftaran-ditutup', !pendaftaranDiizinkan);
  }

  async function muatPengguna() {
    try {
      const response = await fetch('/api/auth/me');
      user = response.ok ? (await response.json()).user : null;
    } catch {
      user = null;
    }
  }

  async function keluar() {
    try {
      await fetch('/api/auth/logout', { method: 'POST' });
    } finally {
      menuPenggunaTerbuka = false;
      dialogPengaturanTerbuka = false;
      user = null;
      window.location.reload();
    }
  }

  function bukaPengaturan() {
    menuPenggunaTerbuka = false;
    muatPengaturan();
    dialogPengaturanTerbuka = true;
  }

  function tutupPengaturan() {
    dialogPengaturanTerbuka = false;
  }

  function tanganiKlikDiLuar(event: MouseEvent) {
    const target = event.target;
    if (target instanceof Element && !target.closest('.menu-pengguna')) menuPenggunaTerbuka = false;
  }

  function tambahKonfirmasiKataSandi(form: HTMLFormElement) {
    if (form.querySelector('[data-konfirmasi-kata-sandi]')) return;
    const kataSandi = form.querySelector<HTMLInputElement>('input[autocomplete="new-password"]');
    if (!kataSandi) return;

    const label = document.createElement('label');
    label.dataset.konfirmasiKataSandi = 'true';
    label.textContent = 'Konfirmasi kata sandi';

    const input = document.createElement('input');
    input.type = 'password';
    input.autocomplete = 'new-password';
    input.minLength = 12;
    input.required = true;
    input.setAttribute('aria-describedby', 'konfirmasi-kata-sandi-pesan');
    label.append(input);

    const pesan = document.createElement('span');
    pesan.id = 'konfirmasi-kata-sandi-pesan';
    pesan.className = 'pesan-konfirmasi-kata-sandi';
    pesan.hidden = true;
    label.append(pesan);

    kataSandi.closest('label')?.after(label);

    const validasi = () => {
      const cocok = input.value === kataSandi.value;
      input.setCustomValidity(cocok || !input.value ? '' : 'Konfirmasi kata sandi tidak sama.');
      pesan.hidden = cocok || !input.value;
      pesan.textContent = 'Konfirmasi kata sandi tidak sama.';
    };

    input.addEventListener('input', validasi);
    kataSandi.addEventListener('input', validasi);
  }

  function siapkanValidasiPendaftaran() {
    const form = document.querySelector<HTMLFormElement>('.layar-autentikasi form');
    if (form) tambahKonfirmasiKataSandi(form);
  }

  onMount(() => {
    muatPengaturan();
    void muatPengguna();

    const observer = new MutationObserver(siapkanValidasiPendaftaran);
    observer.observe(document.body, { childList: true, subtree: true });
    siapkanValidasiPendaftaran();

    const timer = window.setInterval(() => void muatPengguna(), 1500);
    window.addEventListener('click', tanganiKlikDiLuar);
    return () => {
      observer.disconnect();
      window.clearInterval(timer);
      window.removeEventListener('click', tanganiKlikDiLuar);
    };
  });
</script>

<svelte:window onkeydown={(event) => {
  if (event.key === 'Escape') {
    menuPenggunaTerbuka = false;
    dialogPengaturanTerbuka = false;
  }
}} />

{@render children()}

{#if user}
  <div class="menu-pengguna">
    <button
      class="pengguna-atas"
      type="button"
      aria-label="Menu pengguna"
      aria-expanded={menuPenggunaTerbuka}
      onclick={(event) => { event.stopPropagation(); menuPenggunaTerbuka = !menuPenggunaTerbuka; }}
    >
      <span class="avatar">{(user.display_name || user.username).slice(0, 1).toUpperCase()}</span>
      <span class="nama-pengguna">{user.display_name || user.username}</span>
      <span aria-hidden="true">⌄</span>
    </button>

    {#if menuPenggunaTerbuka}
      <div class="daftar-menu-pengguna">
        <button type="button" disabled title="Profil akan tersedia pada tahap berikutnya">Profil</button>
        <button class="berbahaya" type="button" onclick={keluar}>Keluar</button>
      </div>
    {/if}
  </div>

  <button class="tombol-pengaturan" type="button" title="Pengaturan" aria-label="Pengaturan" onclick={bukaPengaturan}>⚙</button>
{/if}

{#if dialogPengaturanTerbuka}
  <div class="lapisan-dialog" role="presentation" onclick={(event) => { if (event.currentTarget === event.target) tutupPengaturan(); }}>
    <section class="dialog-pengaturan" role="dialog" aria-modal="true" aria-labelledby="judul-pengaturan">
      <header>
        <div>
          <p class="eyebrow">PENGATURAN</p>
          <h2 id="judul-pengaturan">Pengaturan</h2>
        </div>
        <button class="tombol-tutup-dialog" type="button" aria-label="Tutup pengaturan" onclick={tutupPengaturan}>×</button>
      </header>

      <div class="isi-pengaturan">
        <label class="item-pengaturan">
          <span>
            <strong>Pengguna boleh mendaftar</strong>
            <small>Izinkan pembuatan akun lokal baru pada perangkat ini.</small>
          </span>
          <input type="checkbox" bind:checked={pendaftaranDiizinkan} onchange={simpanPengaturan} />
        </label>
      </div>
    </section>
  </div>
{/if}

<style>
  :global(.menu-pengguna) { position: fixed; top: 14px; right: 18px; z-index: 40; }
  :global(.pengguna-atas) { display: flex; align-items: center; gap: 8px; min-height: 38px; padding: 4px 9px 4px 5px; border: 1px solid var(--garis); border-radius: 999px; background: var(--permukaan); color: var(--teks); cursor: pointer; font: inherit; box-shadow: 0 3px 14px rgba(24,25,22,.06); }
  :global(.pengguna-atas:hover) { background: var(--permukaan-lembut); }
  :global(.pengguna-atas .avatar) { width: 28px; height: 28px; }
  :global(.nama-pengguna) { max-width: 160px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-weight: 600; }
  :global(.daftar-menu-pengguna) { position: absolute; top: 44px; right: 0; width: 170px; padding: 5px; border: 1px solid var(--garis); border-radius: 10px; background: var(--permukaan); box-shadow: 0 12px 30px rgba(24,25,22,.14); }
  :global(.daftar-menu-pengguna button) { display: block; width: 100%; padding: 9px 10px; border: 0; border-radius: 7px; background: transparent; color: var(--teks); text-align: left; cursor: pointer; font: inherit; }
  :global(.daftar-menu-pengguna button:hover:not(:disabled)) { background: var(--permukaan-lembut); }
  :global(.daftar-menu-pengguna button:disabled) { color: var(--teks-2); cursor: not-allowed; opacity: .6; }
  :global(.daftar-menu-pengguna .berbahaya) { color: #a33a32; }
  :global(.tombol-pengaturan) { position: fixed; left: 18px; bottom: 18px; z-index: 30; width: 38px; height: 38px; border: 1px solid var(--garis); border-radius: 10px; background: var(--permukaan); color: var(--teks-2); cursor: pointer; font-size: 18px; box-shadow: 0 3px 14px rgba(24,25,22,.06); }
  :global(.tombol-pengaturan:hover) { background: var(--permukaan-lembut); color: var(--teks); }
  :global(.akun) { padding-bottom: 62px; }
  :global(.akun .ikon-tombol) { display: none; }
  :global(.aksi-atas .avatar) { visibility: hidden; pointer-events: none; }
  :global(.lapisan-dialog) { position: fixed; inset: 0; z-index: 100; display: grid; place-items: center; padding: clamp(14px, 4vw, 40px); background: rgba(20, 21, 18, .34); }
  :global(.dialog-pengaturan) { width: min(560px, 100%); height: min(520px, calc(100vh - 28px)); max-height: calc(100vh - 28px); display: flex; flex-direction: column; overflow: hidden; border: 1px solid var(--garis); border-radius: 16px; background: var(--permukaan); box-shadow: 0 20px 60px rgba(24,25,22,.2); }
  :global(.dialog-pengaturan > header) { flex: 0 0 auto; display: flex; align-items: flex-start; justify-content: space-between; gap: 20px; padding: 22px 24px 16px; border-bottom: 1px solid var(--garis); }
  :global(.dialog-pengaturan h2) { margin: 2px 0 0; }
  :global(.tombol-tutup-dialog) { width: 32px; height: 32px; border: 0; border-radius: 8px; background: transparent; color: var(--teks-2); cursor: pointer; font-size: 24px; }
  :global(.tombol-tutup-dialog:hover) { background: var(--permukaan-lembut); color: var(--teks); }
  :global(.isi-pengaturan) { flex: 1 1 auto; min-height: 0; overflow: auto; padding: 16px 24px 24px; }
  :global(.item-pengaturan) { display: flex; align-items: center; justify-content: space-between; gap: 24px; padding: 16px; border: 1px solid var(--garis); border-radius: 12px; cursor: pointer; }
  :global(.item-pengaturan span) { display: grid; gap: 4px; }
  :global(.item-pengaturan small) { color: var(--teks-2); line-height: 1.45; }
  :global(.item-pengaturan input) { width: 20px; height: 20px; flex: 0 0 auto; accent-color: currentColor; }
  :global(.pesan-konfirmasi-kata-sandi) { margin-top: 5px; color: #a33a32; font-size: 0.82rem; }
  :global(html[data-pendaftaran-ditutup] .layar-autentikasi .tautan) { display: none; }
  @media (max-width: 760px) {
    :global(.menu-pengguna) { top: 10px; right: 10px; }
    :global(.nama-pengguna) { display: none; }
    :global(.tombol-pengaturan) { left: 10px; bottom: 10px; }
    :global(.dialog-pengaturan) { width: 100%; height: min(560px, calc(100vh - 20px)); border-radius: 14px; }
    :global(.dialog-pengaturan > header) { padding: 18px; }
    :global(.isi-pengaturan) { padding: 12px 18px 18px; }
  }
</style>
