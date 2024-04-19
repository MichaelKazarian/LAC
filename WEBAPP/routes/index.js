const express = require('express');

const router = express.Router();

// use res.render to load up an ejs view file
router.get('/', function(req, res) {  // index page
    res.render('pages/index');
});

router.get('/test', (req, res) => {
    res.send('It works!');
});

module.exports = router;
