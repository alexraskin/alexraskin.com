import { Hono, Context } from 'hono'
import { Chalk } from 'chalk';
import boxen from 'boxen'
import indexHtml from './index.html'

const app = new Hono()

const getCard = () => {
  const chk = new Chalk({ level: 2 });
  const linkTitle = chk.green;
  const pale = chk.gray;
  const highlight = chk.cyan;

  const output = `
              ${chk.green.bold("Alex Raskin")}
      DevOps Engineer Based in America
  
  ${linkTitle("Twitter")}   ${pale("https://")}twitter.com/${highlight("raskin_alex")}

  ${linkTitle("Github")}    ${pale("https://")}github.com/${highlight("alexraskin")}

  ${linkTitle("Bluesky")}  ${pale("https://")}bsky.app/profile/${highlight("alexraskin.bsky.social")}
  
  `;

  return chk.green(
    boxen(chk.white(output), {
      padding: 1,
      margin: 1,
      borderStyle: "single",
    })
  );
};

app.get('/', (c: Context) => {
  const headers = c.req.header()
  if (
    headers["user-agent"].includes("curl" || "HTTPie")
  ) {
    c.header("Content-Type", "text/plain")
    return c.body(getCard(), 200)
  }
  c.header("Content-Type", "text/html")
  return c.html(indexHtml, 200)
})

app.notFound((c) => {
  return c.json({ error: 'Not Found' }, 404)
})

app.onError((err, c) => {
  console.error(`${err}`)
  return c.json({ error: 'Internal Server Error' }, 500)
})

export default app
