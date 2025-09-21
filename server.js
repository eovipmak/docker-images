const express = require("express");
const { exec } = require("child_process");


const app = express();
app.use(express.json());
const PORT = 8000;


function checkCert(domain, port, cb) {
  port = port || 443;
  const cmd = `echo | openssl s_client -servername ${domain} -connect ${domain}:${port} 2>/dev/null | openssl x509 -noout -dates -subject -issuer`;

  exec(cmd, { timeout: 8000 }, (err, stdout, stderr) => {
    if (err) return cb(err);

    const res = {};
    stdout.split("\n").forEach(line => {
      if (line.startsWith("subject=")) res.subject = line.replace("subject=", "").trim();
      if (line.startsWith("issuer=")) res.issuer = line.replace("issuer=", "").trim();
      if (line.startsWith("notBefore=")) res.notBefore = line.replace("notBefore=", "").trim();
      if (line.startsWith("notAfter=")) res.notAfter = line.replace("notAfter=", "").trim();
    });

    if (res.notAfter) {
      const expiry = Date.parse(res.notAfter.replace("GMT", "UTC"));
      if (!isNaN(expiry)) {
        const now = Date.now();
        res.daysRemaining = Math.max(Math.floor((expiry - now) / (1000 * 60 * 60 * 24)), 0);
      }
    }

    cb(null, res);
  });
}



function handleCheck(req, res) {
  let domain, port;
  if (req.method === 'POST') {
    domain = req.body.domain;
    port = req.body.port ? parseInt(req.body.port, 10) : undefined;
  } else {
    domain = req.query.domain;
    port = req.query.port ? parseInt(req.query.port, 10) : undefined;
  }
  if (!domain) return res.status(400).json({ error: "Missing domain" });

  checkCert(domain, port, (err, info) => {
    if (err) return res.status(502).json({ error: err.message });
    res.json({ domain, port: port || 443, ...info });
  });
}


app.get("/", (req, res) => {
  res.json({
    message: "Use POST /check with JSON body: { domain: 'example.com', port: 443 (optional) }"
  });
});
app.post("/check", handleCheck);

app.listen(PORT, '0.0.0.0', () => console.log(`Server running on port ${PORT}`));