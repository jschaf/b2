// Jest Snapshot v1, https://goo.gl/fbAQLP

exports[`PostCompiler should compile a simple post 1`] = `
Object {
  "index.html": "<!doctype html>
<html lang=\\"en\\">
<head>
  <meta charset=\\"utf-8\\">
  <meta name=\\"viewport\\" content=\\"width=device-width, initial-scale=1.0\\">
  <meta name=\\"robots\\" content=\\"index, follow\\">
  <link rel=\\"icon\\" href=\\"/favicon.ico\\">
  <link rel=\\"apple-touch-icon-precomposed\\" href=\\"/favicon-152.png\\">
  <link rel=\\"stylesheet\\" href=\\"/style/main.css\\">
  <script defer src=\\"/instantpage.min.js\\" type=\\"application/javascript\\"></script>
</head>
<body>
  <header>
    <nav class=\\"site-nav\\" role=\\"navigation\\"><a class=\\"site-title\\" href=\\"/\\" title=\\"Home page\\">Joe Schafer</a>
      <ul>
        <li><a href=\\"https://github.com/jschaf\\" title=\\"GitHub page\\">GitHub</a></li>
        <li><a href=\\"https://www.linkedin.com/in/jschaf/\\" title=\\"LinkedIn page\\">LinkedIn</a></li>
      </ul>
    </nav>
  </header>
  <main>
    <div class=\\"main-inner-container\\">
      <article>
        <header><time datetime=\\"2019-10-08\\">October 8, 2019</time>
          <h1><a href=\\"/foo_bar\\">alpha</a></h1>
        </header>
        <section>
          <p>Foo bar.</p>
        </section>
      </article>
    </div>
  </main>
  <footer role=\\"contentinfo\\"><a href=\\"/\\" title=\\"Home page\\">© 2020 Joe Schafer</a></footer>
</body>
</html>",
}
`;
