<script lang="ts">
  function confirmAdult(e: MouseEvent, board: string) {
    const msg =
      board === '/b/'
        ? '18+ WARNING - /b/ is unmoderated (except CP and gore which are strictly moderated). You may encounter NSFW / shocking content. Are you 18+ and want to continue?'
        : '18+ WARNING - /a/ is for hornballs. NSFW content. Are you 18+ and want to continue?';
    if (!confirm(msg)) {
      e.preventDefault();
    }
  }
</script>

<svelte:head>
  <title>216chan</title>
  <link rel="icon" type="image/png" href="/icons/216.png" />
  <meta property="og:type" content="website" />
  <meta property="og:url" content="https://216chan.org/" />
  <meta property="og:title" content="216chan" />
  <meta property="og:description" content="216chan is a Tunisian imageboard. General, Programming, Videogames, STEM, Technology, Adult." />
  <meta property="og:image" content="https://i.imgur.com/xW3Dhf2.png" />
  <meta property="og:site_name" content="216chan" />
  <meta name="twitter:card" content="summary_large_image" />
  <meta name="twitter:title" content="216chan" />
  <meta name="twitter:description" content="216chan is a Tunisian imageboard. General, Programming, Videogames, STEM, Technology, Adult." />
  <meta name="twitter:image" content="https://i.imgur.com/xW3Dhf2.png" />
</svelte:head>

<header class="site-header">
  <div class="header-inner">
    <a href="/" class="brand">
      <img src="/icons/216.png" alt="216chan logo" />
      <span class="brand-text">216chan</span>
    </a>

    <nav class="board-nav" aria-label="Boards">
      <a href="/g/">
        <span class="code">/g/</span>
        <span class="label">General</span>
      </a>
      <a href="/p/">
        <span class="code">/p/</span>
        <span class="label">Programming</span>
      </a>
      <a href="/v/">
        <span class="code">/v/</span>
        <span class="label">Videogames</span>
      </a>
      <a href="/s/">
        <span class="code">/s/</span>
        <span class="label">STEM</span>
      </a>
      <a href="/t/">
        <span class="code">/t/</span>
        <span class="label">Technology</span>
      </a>
      <a href="/a/" on:click={(e) => confirmAdult(e, '/a/')}>
        <span class="code">/a/</span>
        <span class="label">Adult</span>
      </a>
      <a href="/b/" on:click={(e) => confirmAdult(e, '/b/')}>
        <span class="code">/b/</span>
        <span class="label">Random</span>
      </a>
    </nav>
  </div>
</header>

<div class="page-center">
  <div class="wrapper">
    <div class="board-box">
      <div class="topbar">[ 216chan.org ]</div>

      <div class="banner-block">
        <img src="/icons/banner.png" alt="216chan banner" />
      </div>

      <div class="inner">
        <hr />

        <div class="motd">
          <span class="motd-prompt">&gt; motd:</span>
          arch btw
        </div>

        <hr />

        <div class="boards-grid">
          <a href="/g/" class="board-card">
            <div class="board-code">/g/ - General</div>
            <div class="board-desc">General discussions and shitposting about daily life.</div>
          </a>
          <a href="/p/" class="board-card">
            <div class="board-code">/p/ - Programming</div>
            <div class="board-desc">Code, tooling and deploy errors.</div>
          </a>
          <a href="/v/" class="board-card">
            <div class="board-code">/v/ - Videogames</div>
            <div class="board-desc">YAYVIDEOGAMES!!!</div>
          </a>
          <a href="/s/" class="board-card">
            <div class="board-code">/s/ - STEM</div>
            <div class="board-desc">Fucking nerds man. Science, Math, Physics...</div>
          </a>
          <a href="/t/" class="board-card">
            <div class="board-code">/t/ - Technology</div>
            <div class="board-desc">The other nerds. Hardware, PC Builds, etc..</div>
          </a>
          <a href="/a/" class="board-card adult" on:click={(e) => confirmAdult(e, '/a/')}>
            <div class="board-code">/a/ - Adult</div>
            <div class="board-desc">For the hornballs</div>
          </a>
          <a href="/b/" class="board-card adult" on:click={(e) => confirmAdult(e, '/b/')}>
            <div class="board-code">/b/ - Random</div>
            <div class="board-desc">216chan's shithole. You can post whatever you want here (as long as it adheres to rules, retard)</div>
          </a>
        </div>

        <hr />

        <div class="contact-line">
          <strong>Contact:</strong>
          <a href="mailto:contact@216chan.org">contact@216chan.org</a>
          &nbsp;&bull;&nbsp;
          <strong>Discord:</strong>
          <a href="https://discord.gg/tunisian" target="_blank" rel="noopener">discord.gg/tunisian</a>
        </div>
      </div>
    </div>
  </div>
</div>

<footer class="site-footer">
  <div class="footer-links">
    <a href="mailto:contact@216chan.org">Contact</a>
    <span class="footer-sep">&bull;</span>
    <a href="https://discord.gg/tunisian" target="_blank" rel="noopener">Discord</a>
    <span class="footer-sep">&bull;</span>
    <a href="https://github.com/216chan/216-underwork-website" target="_blank" rel="noopener">Source Code</a>
  </div>
  <div class="footer-copy">Copyright &copy; under ground tunisians community. All rights reserved.</div>
  <img class="corner-img" src="/icons/corner.png" alt="216chan corner" />
</footer>

<style>
  .site-header {
    background: #800000;
    border-bottom: 2px solid #600000;
    position: sticky;
    top: 0;
    z-index: 20;
  }

  .header-inner {
    max-width: 1100px;
    margin: 0 auto;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 16px;
    gap: 16px;
  }

  .brand {
    display: flex;
    align-items: center;
    gap: 10px;
    text-decoration: none;
    color: #fff;
    flex-shrink: 0;
  }

  .brand img {
    width: 30px;
    height: 30px;
    display: block;
    image-rendering: pixelated;
  }

  .brand-text {
    font-weight: 700;
    font-size: 20px;
    letter-spacing: 1px;
    line-height: 1;
  }

  .board-nav {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 2px;
    flex: 1;
    flex-wrap: wrap;
  }

  .board-nav a {
    color: #fff;
    text-decoration: none;
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 4px 10px;
    border: 1px solid transparent;
    min-width: 72px;
    line-height: 1;
  }

  .board-nav a:hover {
    background: #ffffee;
    color: #800000;
    border-color: #ffffee;
  }

  .board-nav .code {
    font-weight: 700;
    font-size: 14px;
    letter-spacing: 0.5px;
  }

  .board-nav .label {
    font-size: 10px;
    opacity: 0.95;
    margin-top: 2px;
    letter-spacing: 0.3px;
  }



  .page-center {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 28px 20px 20px;
  }

  .wrapper {
    width: 100%;
    max-width: 720px;
    text-align: center;
  }

  .board-box {
    border: 2px solid #800000;
    background: #ffffee;
    padding: 0 0 20px;
    text-align: center;
  }

  .topbar {
    background: #800000;
    color: #fff;
    font-size: 14px;
    font-weight: bold;
    letter-spacing: 1px;
    padding: 4px 10px;
    text-align: left;
  }

  .banner-block img {
    width: 100%;
    height: auto;
    display: block;
  }

  .inner {
    padding: 0 24px;
  }

  hr {
    border: none;
    border-top: 1px solid #800000;
    margin: 14px 0;
  }

  .motd {
    border: 1px dashed #800000;
    background: #eae9d9;
    padding: 8px 10px;
    font-size: 11px;
    line-height: 1.5;
    color: #800000;
    text-align: left;
  }

  .motd-prompt {
    font-weight: 700;
    color: #cc0000;
  }

  .boards-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 10px;
    text-align: left;
    margin: 6px 0;
  }

  .boards-grid > :last-child:nth-child(3n + 1) {
    grid-column: 2;
  }

  .board-card {
    border: 1px solid #800000;
    background: #eae9d9;
    padding: 12px 10px;
    text-decoration: none;
    color: #800000;
    display: block;
    transition:
      background 0.12s,
      transform 0.08s;
  }

  .board-card:hover {
    background: #fff;
    transform: translateY(-1px);
    box-shadow: 1px 1px 0 #800000;
  }

  .board-card.adult {
    border-style: dashed;
  }

  .board-code {
    font-weight: 700;
    font-size: 13px;
    letter-spacing: 0.6px;
    margin-bottom: 4px;
  }

  .board-desc {
    font-size: 11px;
    line-height: 1.4;
    color: #5a0000;
  }

  .contact-line {
    font-size: 11px;
    color: #800000;
    line-height: 2;
  }

  .contact-line a {
    color: #0000ee;
  }

  .site-footer {
    background-color: #eae9d9;
    border-top: 1px solid #c8c7b4;
    padding: 18px 20px 90px;
    text-align: center;
    position: relative;
  }

  .footer-links {
    font-size: 12px;
    color: #34345c;
    margin-bottom: 10px;
  }

  .footer-links a {
    color: #34345c;
    text-decoration: none;
  }

  .footer-links a:hover {
    text-decoration: underline;
    color: #800000;
  }

  .footer-sep {
    margin: 0 5px;
    color: #800000;
  }

  .footer-copy {
    font-size: 12px;
    color: #800000;
  }

  .corner-img {
    position: absolute;
    bottom: 0;
    right: 0;
    width: 200px;
    height: auto;
    display: block;
    pointer-events: none;
  }

  @media (max-width: 720px) {
    .header-inner {
      flex-wrap: wrap;
      justify-content: center;
    }
    .brand {
      width: 100%;
      justify-content: center;
    }
    .board-nav a {
      min-width: 64px;
      padding: 4px 8px;
    }
    .boards-grid {
      grid-template-columns: repeat(2, 1fr);
    }
    .boards-grid > :last-child:nth-child(3n + 1) {
      grid-column: auto;
    }
    .boards-grid > :last-child:nth-child(2n + 1) {
      grid-column: 1 / span 2;
      max-width: calc(50% - 5px);
      justify-self: center;
      width: 100%;
    }
  }

  @media (max-width: 420px) {
    .boards-grid {
      grid-template-columns: 1fr;
    }
    .inner {
      padding: 0 14px;
    }
  }
</style>
