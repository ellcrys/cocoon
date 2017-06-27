// @flow
import FlexboxBrowser from './flexbox.browser'
import cheerio from 'cheerio'
import {Parser} from 'html-to-react'

class CardML extends FlexboxBrowser {
    
    validTags = ["view"]
    
    // isValidTag checks whether a tag is a valid cml tag
    isValidTag(tag: string): boolean {
        return this.validTags.indexOf(tag) !== -1
    }
    
    // parse takes a card markup, validates it and returns 
    // a react component
    parse(markup: string) : any {
        const $ = (cheerio.load(markup):any)
        const _parse = (select) => {
            $(select).each((i, el) => {
                if (el.type === 'tag') {
                    if (!this.isValidTag(el.tagName)) {
                        throw new Error(`element has invalid tag '${el.tagName}' in ${cheerio.load(el).html()}`)
                    }
                    this.flexify(el)
                }
            })
        }
        _parse('body *')
        var html2React = new Parser()
        return html2React.parse($('body').html())
    }
}

export default CardML;
