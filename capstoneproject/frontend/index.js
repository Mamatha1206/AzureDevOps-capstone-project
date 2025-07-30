// capstoneproject/frontend/index.js
const app = require('./server');
const PORT = process.env.PORT || 3000;

app.listen(PORT, () => {
  console.log(`Frontend server running on http://localhost:${PORT}`);
});
