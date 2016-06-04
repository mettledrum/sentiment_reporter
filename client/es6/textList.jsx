import React from 'react';
import { render } from 'react-dom';

let TextList = React.createClass({
    getInitialState () {
        // connect to web socket hub
        const webSocket = new WebSocket("ws://localhost:12345/ws");

        const handleMessage = this.addText;

        // listen for messages
        webSocket.onmessage = function(evt){
            const text = JSON.parse(evt.data);
            handleMessage(text);
        };

        return {
            texts: []
        };
    },

    addText(newText) {
        console.log(newText);
        newText.id = this.getUniqueID(newText);
        this.setState({ texts: this.state.texts.concat([newText]) });
    },

    getUniqueID(newText) {
        return Date.now().toString() + newText.from;
    },

    render () {
        return (
            <div>
                {this.state.texts.map(function(text) {
                    return <TextItem key={ text.id } from={ text.from } score={ text.score } type={ text.type } content={ text.content }/>;
                })}
            </div>
        );
    }
});

let TextItem = React.createClass({
    render: function() {
        return (
            <div className={ this.props.type }>
                { this.props.from }: { this.props.content }
            </div>
        );
    }
});

module.exports = {
    TextList: TextList
};
