// @flow
import { Component } from 'react';
import EML from '../../eml/eml'
import Cocoon from '../../services/cocoon'
import {InvokeError} from '../../errors/errors'

type State = any
type Props = any

// Base is the base component
class Base extends Component {
    state: State
    props: Props
    eml: EML
    cocoon: Cocoon
    
    constructor(props: {}){
        super(props)
        this.eml = new EML()
        this.cocoon = new Cocoon()
    }
    
    isInvokeError(err: any) {
        return err instanceof InvokeError
    }
}

export default Base