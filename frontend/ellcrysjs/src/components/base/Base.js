// @flow
import { Component } from 'react';
import CardML from '../../cardml/cardml'
import Cocoon from '../../services/cocoon'

// Base is the base component
class Base<PropType,StateType> extends Component<void, PropType, StateType> {
    propTypes: PropType
    state: StateType
    cml: CardML
    cocoon: Cocoon
    
    constructor(props: {}){
        super(props)
        this.cml = new CardML()
        this.cocoon = new Cocoon()
    }
}

export default Base