import FlexboxBrowser from './flexbox.browser'
import parse5 from 'parse5'
import cheerio from 'cheerio'
import {Parser} from 'html-to-react'

class CardML extends FlexboxBrowser {
    
    validTags = ["card"]
    
    // isValidTag checks whether a tag is a valid cml tag
    isValidTag(tag) {
        return this.validTags.indexOf(tag) !== -1
    }
    
    // parseFragment parses a markup and returns a parse5.Document.
    parseFragment(markup) {
        return parse5.parseFragment(markup)
    }
    
    // parse takes a card markup, validates it and returns 
    // a react component
    parse(markup) {
        const $ = cheerio.load(markup)
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
