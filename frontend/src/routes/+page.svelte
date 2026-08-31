<svelte:head>
  <title>216chan - Under Work</title>
  <link rel="icon" type="image/png" href="/icons/216.png" />
  <meta property="og:type" content="website" />
  <meta property="og:url" content="https://216chan.org/" />
  <meta property="og:title" content="216chan — Under Work" />
  <meta property="og:description" content="The site is currently under construction. Stay tuned." />
  <meta property="og:image" content="https://i.imgur.com/xW3Dhf2.png" />
  <meta property="og:site_name" content="216chan" />
  <meta name="twitter:card" content="summary_large_image" />
  <meta name="twitter:title" content="216chan — Under Work" />
  <meta name="twitter:description" content="The site is currently under construction. Stay tuned." />
  <meta name="twitter:image" content="https://i.imgur.com/xW3Dhf2.png" />
</svelte:head>

<div class="page-center">
  <div class="wrapper">
    <div class="board-box">
      <div class="topbar">[ 216chan.org ] &nbsp;—&nbsp; Under Construction</div>

      <div class="banner-block">
        <img src="/icons/banner.png" alt="216chan banner" />
      </div>

      <div class="inner">
        <hr />

        <div class="underwork-title">Underwork<span class="blink">_</span></div>
        <div class="underwork-msg">
          The site is currently under construction.<br />
          Please check back soon.
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
  .page-center {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 40px 20px 20px;
  }

  .wrapper {
    width: 100%;
    max-width: 680px;
    text-align: center;
  }

  .board-box {
    border: 2px solid #800000;
    background: #ffffee;
    padding: 0 0 20px;
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
    margin: 12px 0;
  }

  .underwork-title {
    font-size: 30px;
    font-weight: bold;
    color: #cc0000;
    font-family: "Times New Roman", Times, serif;
    text-transform: uppercase;
    letter-spacing: 5px;
    margin: 18px 0 6px;
  }

  .underwork-msg {
    font-size: 13px;
    color: #800000;
    margin-bottom: 10px;
    line-height: 1.6;
  }

  .blink {
    animation: blink-anim 1s step-end infinite;
  }

  @keyframes blink-anim {
    50% {
      opacity: 0;
    }
  }

  .contact-line {
    font-size: 12px;
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
  }
</style>
