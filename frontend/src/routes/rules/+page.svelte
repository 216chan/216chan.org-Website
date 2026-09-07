<svelte:head>
  <title>Rules - 216chan</title>
  <link rel="icon" type="image/png" href="/icons/216.png" />
  <meta name="description" content="216chan rules - global and board rules. Tunisian imageboard." />
  <meta property="og:title" content="Rules - 216chan" />
  <meta property="og:description" content="Read the rules for 216chan before posting." />
  <meta property="og:image" content="https://frontend-production-aff6.up.railway.app/icons/embed.png" />
  <meta name="theme-color" content="#800000" />
</svelte:head>

<header class="site-header">
  <div class="header-inner">
    <a href="/" class="brand"><img src="/icons/216.png" alt="logo" /><span class="brand-text">216chan</span></a>
    <nav class="board-nav">
      {#each [['g','General'],['p','Programming'],['v','Videogames'],['s','STEM'],['t','Technology'],['a','Adult'],['b','Random']] as [slug,label]}
        <a href="/{slug}/"><span class="code">/{slug}/</span><span class="label">{label}</span></a>
      {/each}
    </nav>
  </div>
</header>

<div class="page-center">
  <div class="wrapper">
    <div class="board-box">
      <div class="topbar">Rules</div>
      <div class="inner">
        <div class="rules-intro">216chan is anonymous. Tunisian law always applies. Break the rules and your post will be removed. No warnings for serious stuff.</div>
        <hr />
        <div class="global">
          <div class="box-title">Global</div>
          <div class="box-body">
            <p>1. No CP. No sexual content with minors. Zero tolerance.</p>
            <p>2. No gore or shock with real people.</p>
            <p>3. No dox. No personal info without consent.</p>
            <p>4. No spam, flood, or ads. No crypto scam.</p>
            <p>5. Use the right board. Off topic posts will be moved or removed.</p>
          </div>
        </div>
        <div class="boards">
          <div class="box-title">By board</div>
          <div class="box-body">
            <p>Board rules are in the pinned thread on each board. Check the top thread on <a href="/g/">/g/</a> <a href="/p/">/p/</a> <a href="/v/">/v/</a> <a href="/s/">/s/</a> <a href="/t/">/t/</a> <a href="/a/">/a/</a> <a href="/b/">/b/</a></p>
            <p>Mods can edit pinned threads via the mod panel. No hardcode here.</p>
          </div>
        </div>
        <hr />
        <div class="contact-line">Questions: <a href="mailto:contact@216chan.org">contact@216chan.org</a> - <a href="/">back to home</a></div>
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
  .board-nav a:hover{ background:#ffffee; color:#800000; border-color:#ffffee; }
  .board-nav .code{ font-weight:700; font-size:12px; color:#ffaaaa; }
  .board-nav .label{ font-size:9px; opacity:.9; margin-top:2px; }
  .page-center{ flex:1; display:flex; justify-content:center; padding:20px 12px 16px; }
  .wrapper{ width:100%; max-width:720px; text-align:center; }
  .board-box{ border:2px solid #800000; background:#ffffee; padding:0 0 20px; text-align:center; border-radius:10px; overflow:hidden; box-shadow:0 2px 10px rgba(128,0,0,.18); }
  .topbar{ background:linear-gradient(to right,#800000,#cc3333); color:#fff; font-size:13px; font-weight:bold; letter-spacing:1px; padding:6px 10px; text-align:center; }
  .inner{ padding:0 16px; text-align:left; }
  hr{ border:none; border-top:1px solid #800000; margin:12px 0; }
  .rules-intro{ text-align:center; font-size:12px; color:#800000; padding-top:12px; }
  .box-title{ background:linear-gradient(to right,#800000,#cc3333); color:#fff; font-weight:bold; font-size:13px; padding:7px 12px; border-radius:9px 9px 0 0; }
  .box-body{ background:#fff; border:1px solid #a00000; border-top:none; padding:10px 14px; font-size:12px; color:#222; line-height:1.7; border-radius:0 0 9px 9px; margin-bottom:12px; }
  .box-body p{ margin:4px 0; }
  .global,.boards{ text-align:left; margin:6px 0; border-radius:10px; overflow:hidden; border:1px solid #a00000; box-shadow:0 2px 8px rgba(128,0,0,.15); }
  .global .box-body, .boards .box-body{ border:none; margin-bottom:0; }
  .contact-line{ font-size:11px; color:#800000; text-align:center; }
  .contact-line a{ color:#0000ee; }
  .site-footer{ background:#eae9d9; border-top:1px solid #c8c7b4; padding:14px 0 20px; text-align:center; }
  .footer-nav{ display:flex; justify-content:center; gap:6px; flex-wrap:wrap; margin-bottom:8px; }
  .footer-nav a{ border:1px solid #aaa; background:#f0f0e8; color:#34345c; text-decoration:none; font-size:12px; padding:4px 10px; }
  .footer-copy{ font-size:11px; color:#800000; }
  @media(max-width:600px){ .header-inner{ flex-direction:column; align-items:stretch; padding:6px 10px; gap:4px; } .brand{ justify-content:center; } .board-nav a{ min-width:44px; padding:5px 6px; flex:1; } .inner{ padding:0 10px; } }
</style>