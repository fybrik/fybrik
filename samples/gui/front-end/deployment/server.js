'use strict';
    
const express = require('express');
const path = require('path');
    
// Constants
const PORT = 3000;
    
// App
const app = express();
app.use(express.static(path.join(__dirname, 'build')));
app.get('/*', function (req, res) {
res.sendFile(path.join(__dirname, 'build', 'index.html'));
});
    
app.listen(PORT);
