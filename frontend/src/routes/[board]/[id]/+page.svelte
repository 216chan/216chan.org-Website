<script lang="ts">
  import type { PageData } from './$types';
  export let data: PageData;
  $: board = data.board;
  $: id = data.id;
  $: info = data.data;
  $: error = data.error;

  let body = '';
  let file: FileList | null = null;
  let submitting = false;
  let msg = '';

  function apiBase(): string {
    try {
      const v = import.meta.env?.VITE_API_URL as string | undefined;
      if (v) return v.replace(/\/api\/?$/, '');
    } catch {}
    return '';
  }
  function imgUrl(u: string | null | undefined): string {
    if (!u) return '';
    if (u.startsWith('http')) return u;
    const base = apiBase();
    if (base) return base + u;
    return u;
  }
  function fmtTime(s: string): string {
    try { return new Date(s).toLocaleString('en-GB', { dateStyle:'short', timeStyle:'short'}); } catch { return s; }
  }
  async function postReply() {
    submitting=true; msg='';
    const fd=new FormData();
    fd.append('body', body);
    if (file && file[0]) fd.append('image', file[0]);
    const base=apiBase();
    const url = base ? `${base}/api/${board}/threads/${id}/replies` : `/api/${board}/threads/${id}/replies`;
    try{
      const r=await fetch(url,{method:'POST', body:fd});
      const j=await r.json().catch(()=>({}));
      if(!r.ok){ msg=j.error||`Error ${r.status}`; return; }
      location.reload();
    }catch(e){ msg=String(e);} finally{submitting=false;}
  }
  function goPage(p:number){ window.location.search=`?page=${p}`; }
  function toggleImg(e: MouseEvent){ (e.currentTarget as HTMLImageElement).classList.toggle('expanded'); }
</script>

<svelte:head>
  <title>/{board}/{id} - 216chan</title>
  <link rel="icon" type="image/png" href="/icons/216.png" />
  <meta name="description" content="216chan /{board}/ thread {id} - Tunisian imageboard." />
  <meta property="og:title" content="/{board}/ #{id} - 216chan" />
  <meta property="og:description" content="View thread on /{board}/ - 216chan." />
  <meta property="og:image" content="https://frontend-production-aff6.up.railway.app/icons/embed.png" />
  <meta property="og:type" content="website" />
  <meta name="theme-color" content="#800000" />
</svelte:head>

<header class="site-header">
  <div class="header-inner">
    <a href="/" class="brand"><img src="/icons/216.png" alt="logo"/><span class="brand-text">216chan</span></a>
    <nav class="board-nav">
      {#each [['g','General'],['p','Programming'],['v','Videogames'],['s','STEM'],['t','Technology'],['a','Adult'],['b','Random']] as [slug,label]}
        <a href="/{slug}/" class:active={slug===board}><span class="code">/{slug}/</span><span class="label">{label}</span></a>
      {/each}
    </nav>
  </div>
</header>

<div class="page-center">
  <div class="wrapper">
    <div class="board-box">
      <div class="topbar"><a href="/{board}/" style="color:#fff; text-decoration:none;">/{board}/</a> - Thread #{id}</div>
      <div class="inner">
        {#if error && !info}
          <div class="err">Failed: {error}</div>
          <div style="text-align:center; margin-top:12px;"><a href="/{board}/" class="back">- Back to /{board}/</a></div>
        {:else if info}
          <div style="text-align:center; padding:8px 0 0;"><a href="/{board}/" class="back">- Back to /{board}/ - {info.board.name}</a></div>
          <hr/>
          {#if info.thread.is_pinned}<div class="pinned-label">[Pinned]</div>{/if}
          {#if info.thread.is_locked}<div class="locked-label">[Locked] - replies disabled</div>{/if}

          <div class="post op">
            <div class="post-head">
              <span class="name">anonymous</span>
              <span class="country">[{info.op.country_code}]</span>
              <span class="time">{fmtTime(info.op.created_at)}</span>
              <span class="post-link">No.{info.op.id}</span>
              <span class="op-tag">[OP]</span>
            </div>
            {#if info.op.image_url}
              <div class="file-info">{info.op.image_name ?? 'image'}</div>
              <img class="post-img" src={imgUrl(info.op.image_url)} alt={info.op.image_name ?? 'image'} on:click={toggleImg} />
            {/if}
            <div class="post-body">{info.op.body}</div>
            <div class="clear"></div>
          </div>

          <hr/>

          {#if info.replies.length===0}
            <div class="empty">No replies yet.</div>
          {:else}
            {#each info.replies as r}
              <div class="post reply">
                <div class="post-head">
                  <span class="name">anonymous</span>
                  <span class="country">[{r.country_code}]</span>
                  <span class="time">{fmtTime(r.created_at)}</span>
                  <span class="post-link">No.{r.id}</span>
                </div>
                {#if r.image_url}
                  <div class="file-info">{r.image_name ?? 'image'}</div>
                  <img class="post-img small" src={imgUrl(r.image_url)} alt={r.image_name ?? 'image'} loading="lazy" on:click={toggleImg} />
                {/if}
                <div class="post-body">{r.body}</div>
                <div class="clear"></div>
              </div>
            {/each}
          {/if}

          {#if info.pages>1}
            <div class="pages">
              {#each Array(info.pages) as _, i}
                <a href="/{board}/{id}?page={i+1}" class:active={info.page===i+1} on:click|preventDefault={()=>goPage(i+1)}>{i+1}</a>
              {/each}
            </div>
            <div class="pages-info">Page {info.page} / {info.pages} - {info.total} replies</div>
          {/if}

          <hr/>

          {#if !info.thread.is_locked}
            <div class="form-box">
              <div class="form-title">Post a reply</div>
              <div class="form-body">
                <textarea bind:value={body} placeholder="Body (or leave empty if posting image)" rows="4" maxlength={5000}></textarea>
                <div class="form-row">
                  <input type="file" accept=".jpg,.jpeg,.png,.gif,.webp" bind:files={file} />
                  <button on:click={postReply} disabled={submitting}>{submitting?'Posting...':'Reply'}</button>
                </div>
                {#if msg}<div class="form-msg">{msg}</div>{/if}
                <div class="form-hint">Max 8MB. Images: jpg png gif webp.</div>
              </div>
            </div>
          {:else}
            <div class="locked-msg">This thread is locked.</div>
          {/if}

        {:else}
          <div class="loading">Loading...</div>
        {/if}
      </div>
    </div>
  </div>
</div>

<footer class="site-footer">
  <nav class="footer-nav">
    <a href="/">Home</a>
    <a href="/rules">Rules</a>
    <a href="mailto:contact@216chan.org">Contact</a>
    <a href="https://discord.gg/tunisian" target="_blank" rel="noopener">Discord</a>
    <a href="https://github.com/216chan/216-underwork-website" target="_blank" rel="noopener">Source</a>
  </nav>
  <div class="footer-copy">Copyright © Underground Tunisians Community. All rights reserved.</div>
</footer>

<style>
  .site-header{ background:#800000; border-bottom:2px solid #600000; position:sticky; top:0; z-index:20; }
  .header-inner{ width:100%; display:flex; align-items:center; padding:6px 12px; min-height:46px; gap:8px; }
  .brand{ display:flex; align-items:center; gap:6px; text-decoration:none; color:#fff; flex-shrink:0; white-space:nowrap; }
  .brand img{ width:28px; height:28px; display:block; image-rendering:pixelated; }
  .brand-text{ font-weight:700; font-size:18px; letter-spacing:1px; line-height:1; }
  .board-nav{ flex:1; display:flex; align-items:center; justify-content:center; gap:0; flex-wrap:wrap; }
  .board-nav a{ color:#fff; text-decoration:none; display:flex; flex-direction:column; align-items:center; padding:4px 8px; border:1px solid transparent; min-width:52px; line-height:1; }
  .board-nav a:hover,.board-nav a.active{ background:#ffffee; color:#800000; border-color:#ffffee; }
  .board-nav .code{ font-weight:700; font-size:12px; color:#ffaaaa; }
  .board-nav a.active .code,.board-nav a:hover .code{ color:#800000; }
  .board-nav .label{ font-size:9px; opacity:.9; margin-top:2px; }
  .page-center{ flex:1; display:flex; justify-content:center; padding:20px 12px 16px; }
  .wrapper{ width:100%; max-width:860px; text-align:center; }
  .board-box{ border:2px solid #800000; background:#ffffee; padding:0 0 20px; text-align:center; border-radius:10px; overflow:hidden; box-shadow:0 2px 10px rgba(128,0,0,.18); }
  .topbar{ background:linear-gradient(to right,#800000,#cc3333); color:#fff; font-size:13px; font-weight:bold; letter-spacing:1px; padding:6px 10px; text-align:center; }
  .inner{ padding:0 12px; text-align:left; }
  hr{ border:none; border-top:1px solid #d9bfb7; margin:12px 0; }
  .err{ background:#ffe0e0; border:1px solid #c00; color:#800000; padding:10px; text-align:center; }
  .back{ color:#34345c; text-decoration:none; font-size:12px; font-weight:700; }
  .back:hover{ color:#d00; text-decoration:underline; }
  .pinned-label{ background:#ffe; border:1px dashed #800000; color:#800000; display:inline-block; padding:1px 4px; font-size:11px; font-weight:700; margin-bottom:6px; }
  .locked-label{ background:#eee; border:1px solid #800000; color:#800000; display:inline-block; padding:1px 4px; font-size:11px; font-weight:700; margin-bottom:6px; margin-left:6px; }
  .post{ padding:8px 0; overflow:hidden; }
  .post.op{ margin-bottom:6px; }
  .post.reply{ background:#f0e0d6; border:1px solid #d9bfb7; padding:8px 10px; margin:6px 0; border-radius:6px; }
  .post-head{ font-size:11px; color:#800000; display:flex; gap:6px; align-items:center; flex-wrap:wrap; margin-bottom:6px; }
  .name{ color:#117743; font-weight:700; }
  .country{ font-weight:700; color:#34345c; font-size:11px; }
  .time{ color:#666; font-size:11px; }
  .post-link{ color:#800000; text-decoration:none; }
  .post-link:hover{ color:#d00; text-decoration:underline; }
  .op-tag{ color:#800000; font-weight:700; font-size:11px; }
  .file-info{ font-size:11px; color:#34345c; margin:4px 0; }
  .post-img{ float:left; width:250px; height:auto; max-height:250px; object-fit:contain; border:1px solid #d9bfb7; margin:0 14px 8px 0; display:block; cursor:zoom-in; background:#fff; }
  .post-img.small{ width:180px; max-height:180px; }
  .post-img.expanded{ width:auto; height:auto; max-width:100%; max-height:none; float:none; cursor:zoom-out; margin:6px 0; background:transparent; }
  .post-body{ font-size:13px; color:#800000; line-height:1.5; white-space:pre-wrap; word-break:break-word; }
  .clear{ clear:both; }
  .empty{ text-align:center; color:#800000; padding:20px 0; font-style:italic; }
  .pages{ display:flex; justify-content:center; gap:4px; flex-wrap:wrap; margin:14px 0 6px; }
  .pages a{ border:1px solid #d9bfb7; background:#f0e0d6; color:#800000; text-decoration:none; font-size:12px; padding:3px 8px; min-width:24px; text-align:center; }
  .pages a.active,.pages a:hover{ background:#800000; color:#fff; border-color:#800000; }
  .pages-info{ text-align:center; font-size:11px; color:#800000; }
  .form-box{ border:1px solid #a00000; background:#f0e0d6; border-radius:10px; overflow:hidden; box-shadow:0 2px 8px rgba(128,0,0,.15); }
  .form-title{ background:linear-gradient(to right,#800000,#cc3333); color:#fff; font-weight:bold; font-size:13px; padding:7px 12px; border-radius:9px 9px 0 0; }
  .form-body{ padding:12px; display:flex; flex-direction:column; gap:8px; background:#ffffee; border-radius:0 0 9px 9px; }
  textarea{ width:100%; border:1px solid #aaa; background:#fff; padding:6px; font-family:inherit; font-size:13px; resize:vertical; border-radius:4px; }
  .form-row{ display:flex; gap:8px; align-items:center; flex-wrap:wrap; }
  input[type=file]{ font-size:12px; }
  button{ background:#800000; color:#fff; border:1px solid #600000; padding:6px 14px; cursor:pointer; font-weight:700; font-size:12px; border-radius:4px; }
  button:hover{ background:#800000; color:#fff; }
  button:disabled{ opacity:.6; cursor:wait; }
  .form-msg{ color:#c00; font-size:12px; font-weight:700; }
  .form-hint{ font-size:11px; color:#800000; }
  .loading{ text-align:center; padding:20px; color:#800000; font-size:13px; }
  .locked-msg{ text-align:center; color:#800000; font-weight:700; padding:12px; border:1px solid #c00; background:#ffe0e0; }
  .site-footer{ background:#eae9d9; border-top:1px solid #c8c7b4; padding:14px 0 20px; text-align:center; }
  .footer-nav{ display:flex; justify-content:center; gap:6px; flex-wrap:wrap; margin-bottom:8px; }
  .footer-nav a{ border:1px solid #aaa; background:#f0f0e8; color:#34345c; text-decoration:none; font-size:12px; padding:4px 10px; }
  .footer-copy{ font-size:11px; color:#800000; }
  @media(max-width:600px){
    .header-inner{ flex-direction:column; align-items:stretch; padding:6px 10px; gap:4px; }
    .brand{ justify-content:center; }
    .board-nav a{ min-width:44px; padding:5px 6px; flex:1; }
    .post-img{ max-width:100%; float:none; margin:6px 0; }
    .inner{ padding:0 8px; }
  }
</style>