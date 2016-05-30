// create the websocket connection
var socket = new WebSocket("ws://localhost:12345/ws");

var $sentiment = $('#sentiment');

socket.onmessage = function(evt){
  var data = JSON.parse(evt.data);

  console.log(data);
  var entry = '<div class="'+ data.type +'">' + data.content + '</div>';
  $sentiment.append(entry);
};
