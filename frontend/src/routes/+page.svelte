<script lang="ts">
  import { onMount } from 'svelte';
  import type { PageData } from './$types';

  export let data: PageData;

  $: stats = data.stats;

  function formatNumber(n: number): string {
    return n.toLocaleString('en-US');
  }

  const cornerImages = ['/icons/corner.png', '/icons/corner2.png', '/icons/corner3.png'];
  let cornerImg: string | null = null;

  onMount(() => {
    cornerImg = cornerImages[Math.floor(Math.random() * cornerImages.length)];
  });

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
  <meta property="og:image" content="https://frontend-production-aff6.up.railway.app/icons/banner.png" />
  <meta property="og:image:width" content="1200" />
  <meta property="og:image:height" content="630" />
  <meta property="og:site_name" content="216chan" />
  <meta name="twitter:card" content="summary_large_image" />
  <meta name="twitter:title" content="216chan" />
  <meta name="twitter:description" content="216chan is a Tunisian imageboard. General, Programming, Videogames, STEM, Technology, Adult." />
  <meta name="twitter:image" content="https://frontend-production-aff6.up.railway.app/icons/banner.png" />
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
        <img class="banner-watermark" src="/icons/216chan-logo.png" alt="" aria-hidden="true" />
      </div>

      <div class="inner">
        <hr />

        <div class="welcome-block">
          <div class="welcome-text glitch" data-text="Welcome to 216chan">Welcome to 216chan</div>
          <div class="welcome-sub">Tunisian imageboard — anonymous, unfiltered, community-run.</div>
        </div>

        <hr />

        <div class="about-box">
          <div class="about-title">
            What is 216chan?
            <img class="teto" src="/icons/teto.png" alt="teto" />
          </div>
          <div class="about-body">
            <p>216chan is a Tunisian imageboard, a place where people post images and discuss whatever they want, anonymously.</p>
            <p><a href="/g/" class="slug-link"><span class="slug">/g/</span></a> : general talk, daily life, shitposting</p>
            <p><a href="/p/" class="slug-link"><span class="slug">/p/</span></a> : programming, dev tools, deploy pain</p>
            <p><a href="/v/" class="slug-link"><span class="slug">/v/</span></a> : video games, reviews, game dev</p>
            <p><a href="/s/" class="slug-link"><span class="slug">/s/</span></a> : science, math, physics for nerds</p>
            <p><a href="/t/" class="slug-link"><span class="slug">/t/</span></a> : tech, hardware, PC builds</p>
            <p><a href="/a/" class="slug-link"><span class="slug">/a/</span></a> : 18+ adult board</p>
            <p><a href="/b/" class="slug-link"><span class="slug">/b/</span></a> : random, anything goes (within rules)</p>
          </div>
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
  <div class="footer-stats-header">Stats</div>
  <div class="footer-stats-row">
    <span><strong>Total Posts:</strong> {formatNumber(stats.total_posts)}</span>
    <span><strong>Current Users:</strong> {formatNumber(stats.current_users)}</span>
    <span><strong>Active Content:</strong> {stats.active_content}</span>
  </div>

  <nav class="footer-nav" aria-label="Site links">
    <a href="/">Home</a>
    <a href="/news">News</a>
    <a href="/blog">Blog</a>
    <a href="/faq">FAQ</a>
    <a href="/rules">Rules</a>
    <a href="mailto:contact@216chan.org">Contact</a>
    <a href="https://discord.gg/tunisian" target="_blank" rel="noopener">Discord</a>
    <a href="https://github.com/216chan/216-underwork-website" target="_blank" rel="noopener">Source</a>
  </nav>

  <div class="footer-copy">Copyright &copy; Underground Tunisians Community. All rights reserved.</div>
  {#if cornerImg}
    <img class="corner-img" src={cornerImg} alt="216chan corner" />
  {/if}
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
    width: 100%;
    position: relative;
    display: flex;
    align-items: center;
    padding: 6px 16px;
    min-height: 46px;
  }

  .brand {
    display: flex;
    align-items: center;
    gap: 8px;
    text-decoration: none;
    color: #fff;
    flex-shrink: 0;
    order: -1;
    margin-left: 0;
    z-index: 1;
  }

  .brand img {
    width: 28px;
    height: 28px;
    display: block;
    image-rendering: pixelated;
  }

  .brand-text {
    font-weight: 700;
    font-size: 18px;
    letter-spacing: 1px;
    line-height: 1;
  }

  .board-nav {
    position: absolute;
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 2px;
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
    min-width: 68px;
    line-height: 1;
  }

  .board-nav a:hover {
    background: #ffffee;
    color: #800000;
    border-color: #ffffee;
  }

  .board-nav .code {
    font-weight: 700;
    font-size: 13px;
    letter-spacing: 0.5px;
    color: #ffaaaa;
  }

  .board-nav .label {
    font-size: 10px;
    opacity: 0.9;
    margin-top: 2px;
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
    border-radius: 10px;
    overflow: hidden;
    box-shadow: 0 2px 10px rgba(128,0,0,0.18);
  }

  .topbar {
    background: linear-gradient(to right, #800000, #cc3333);
    color: #fff;
    font-size: 13px;
    font-weight: bold;
    letter-spacing: 1px;
    padding: 7px 10px;
    position: relative;
  }

  .banner-block {
    position: relative;
    line-height: 0;
  }

  .banner-block img:first-child {
    width: 100%;
    height: auto;
    display: block;
  }

  .banner-watermark {
    position: absolute;
    bottom: 10px;
    left: 10px;
    height: 40px;
    width: auto;
    opacity: 0.75;
    pointer-events: none;
  }

  .inner {
    padding: 0 24px;
  }

  hr {
    border: none;
    border-top: 1px solid #800000;
    margin: 14px 0;
  }

  .slug {
    color: #cc0000;
    font-weight: 700;
  }

  .slug-link {
    text-decoration: none;
  }

  .slug-link:hover .slug {
    text-decoration: underline;
  }

  .contact-line {
    font-size: 11px;
    color: #800000;
    line-height: 2;
  }

  .contact-line a {
    color: #0000ee;
  }

  .about-box {
    text-align: left;
    margin: 6px 0;
    border-radius: 10px;
    overflow: visible;
    border: 1px solid #a00000;
    box-shadow: 0 2px 8px rgba(128,0,0,0.15);
  }

  .about-title {
    background: linear-gradient(to right, #800000, #cc3333);
    color: #fff;
    font-weight: bold;
    font-size: 13px;
    padding: 7px 12px;
    border-radius: 9px 9px 0 0;
    position: relative;
    overflow: visible;
  }

  .teto {
    position: absolute;
    bottom: 85%;
    right: -5px;
    height: 70px;
    width: auto;
    pointer-events: none;
    image-rendering: auto;
  }

  .about-body {
    background: transparent;
    padding: 10px 14px;
    font-size: 12px;
    color: #222;
    line-height: 1.8;
    border-radius: 0 0 9px 9px;
  }

  .about-body p {
    margin-bottom: 4px;
  }

  .welcome-block {
    padding: 18px 0 10px;
    text-align: center;
  }

  .welcome-sub {
    font-size: 11px;
    color: #5a0000;
    margin-top: 6px;
    letter-spacing: 0.3px;
  }

  .glitch {
    position: relative;
    display: inline-block;
    font-size: 26px;
    font-weight: 700;
    color: #800000;
    letter-spacing: 2px;
  }

  .glitch::before,
  .glitch::after {
    content: attr(data-text);
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    overflow: hidden;
    font-size: 26px;
    font-weight: 700;
    letter-spacing: 2px;
  }

  .glitch::before {
    color: #cc0000;
    text-shadow: 2px 0 #cc0000;
    clip-path: inset(0 0 60% 0);
    animation: glitch-top 3.5s infinite steps(1);
  }

  .glitch::after {
    color: #5a0000;
    text-shadow: -2px 0 #5a0000;
    clip-path: inset(60% 0 0 0);
    animation: glitch-bottom 3.5s infinite steps(1);
  }

  @keyframes glitch-top {
    0%   { clip-path: inset(0 0 80% 0);   transform: translate(-2px, 0); }
    10%  { clip-path: inset(10% 0 70% 0); transform: translate( 2px, 0); }
    20%  { clip-path: inset(20% 0 50% 0); transform: translate(-1px, 0); }
    30%  { clip-path: inset(0 0 90% 0);   transform: translate( 3px, 0); }
    40%  { clip-path: inset(30% 0 60% 0); transform: translate(-2px, 0); }
    50%  { clip-path: inset(5% 0 75% 0);  transform: translate( 0,   0); }
    60%  { clip-path: inset(0 0 80% 0);   transform: translate( 2px, 0); }
    70%  { clip-path: inset(15% 0 65% 0); transform: translate(-3px, 0); }
    80%  { clip-path: inset(0 0 85% 0);   transform: translate( 1px, 0); }
    90%  { clip-path: inset(25% 0 55% 0); transform: translate(-1px, 0); }
    100% { clip-path: inset(0 0 80% 0);   transform: translate( 2px, 0); }
  }

  @keyframes glitch-bottom {
    0%   { clip-path: inset(70% 0 0 0); transform: translate( 2px, 0); }
    10%  { clip-path: inset(60% 0 0 0); transform: translate(-2px, 0); }
    20%  { clip-path: inset(75% 0 0 0); transform: translate( 1px, 0); }
    30%  { clip-path: inset(55% 0 0 0); transform: translate(-3px, 0); }
    40%  { clip-path: inset(80% 0 0 0); transform: translate( 2px, 0); }
    50%  { clip-path: inset(65% 0 0 0); transform: translate( 0,   0); }
    60%  { clip-path: inset(70% 0 0 0); transform: translate(-2px, 0); }
    70%  { clip-path: inset(50% 0 0 0); transform: translate( 3px, 0); }
    80%  { clip-path: inset(75% 0 0 0); transform: translate(-1px, 0); }
    90%  { clip-path: inset(60% 0 0 0); transform: translate( 1px, 0); }
    100% { clip-path: inset(70% 0 0 0); transform: translate(-2px, 0); }
  }

  .site-footer {
    background-color: #eae9d9;
    border-top: 1px solid #c8c7b4;
    padding: 0 0 90px;
    text-align: center;
    position: relative;
  }

  .footer-stats-header {
    background: linear-gradient(to bottom, #f0c8b0, #e8a898);
    border: 1px solid #c8876c;
    font-size: 13px;
    font-weight: bold;
    color: #800000;
    text-align: left;
    padding: 4px 10px;
    margin: 14px auto 0;
    max-width: 900px;
  }

  .footer-stats-row {
    background: linear-gradient(to bottom, #f8e0d8, #f0d0c0);
    border: 1px solid #c8876c;
    border-top: none;
    display: flex;
    justify-content: space-around;
    align-items: center;
    padding: 8px 16px;
    font-size: 12px;
    color: #000;
    margin: 0 auto 10px;
    max-width: 900px;
    flex-wrap: wrap;
    gap: 6px;
  }

  .footer-stats-row strong {
    color: #800000;
  }

  .footer-nav {
    display: flex;
    justify-content: center;
    flex-wrap: wrap;
    margin: 8px auto 12px;
  }

  .footer-nav a {
    display: inline-block;
    border: 1px solid #aaa;
    background: #f0f0e8;
    color: #34345c;
    text-decoration: none;
    font-size: 12px;
    padding: 4px 12px;
    margin: 2px;
    line-height: 1.5;
  }

  .footer-nav a:hover {
    background: #ffffee;
    color: #800000;
    border-color: #800000;
  }

  .footer-copy {
    font-size: 11px;
    color: #800000;
    padding: 0 20px 10px;
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
      gap: 6px;
    }

    .brand {
      order: -1;
      margin-left: 0;
      width: 100%;
      justify-content: flex-start;
    }

    .board-nav {
      width: 100%;
      justify-content: center;
    }

    .board-nav a {
      min-width: 60px;
      padding: 4px 8px;
    }
  }

  @media (max-width: 420px) {
    .inner {
      padding: 0 14px;
    }
  }
</style>
