// @flow
import Browser from './browser'
import cheerio from 'cheerio'
import {Parser} from 'html-to-react'

/**
 * EML defines a module for interpreting and verifying the Ellcrys Markup Language
 * @class EML
 * @extends {Browser}
 */
class EML extends Browser {
    
    validTags = ["view"]
    
    // isValidTag checks whether a tag is a valid cml tag
    isValidTag(tag: string): boolean {
    return this.validTags.indexOf(tag) !== -1
    }
    
    // parse takes a card markup, validates it and returns 
    // a react component
    parse(markup: string) : any {
        const $: any = (cheerio.load(markup):any)
        const _parse = (select) => {
            $(select).each((i: number, el: cheerio.CheerioElement ) => {
                if (el.type === 'tag') {
                    if (!this.isValidTag(el.tagName)) {
                        throw new Error(`element has invalid tag '${el.tagName}' in ${cheerio.load(el).html()}`)
                    }
                    this.apply(el)
                }
            })
        }
        _parse('body *')
        var html2React = new Parser()
        return html2React.parse($('body').html())
    }
}

export default EML;
