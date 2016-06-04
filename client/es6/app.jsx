import React from 'react';
import { render } from 'react-dom';

import { TextList } from './textList.jsx';

let App = React.createClass({
    render () {
        return (
            <div>
                <TextList />
            </div>
        );
    }
});

render(<App/>, document.getElementById('sentiment'));
