//SPDX-License-Identifier: Apache-2.0

var produce = require('./invoker.js');

module.exports = function(app){

  app.get('/get_produce/:id', function(req, res){
    produce.get_produce(req, res);
  });
  app.get('/add_produce/:produce', function(req, res){
    produce.add_produce(req, res);
  });
  app.get('/get_all_produce', function(req, res){
    produce.get_all_produce(req, res);
  });
  app.get('/change_holder/:holder', function(req, res){
    produce.change_holder(req, res);
  });
}
