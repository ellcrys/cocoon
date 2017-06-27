import {AddToInlineStyle} from '../../util/util'
import cheerio from 'cheerio'

export type PaddingApplier = {
    applyPadding: ($el: cheerio.CheerioElement, defaultValue?: string) => void,
    applyPaddingTop: ($el: cheerio.CheerioElement, defaultValue?: string) => void,
    applyPaddingRight: ($el: cheerio.CheerioElement, defaultValue?: string) => void,
    applyPaddingBottom: ($el: cheerio.CheerioElement, defaultValue?: string) => void,
    applyPaddingLeft: ($el: cheerio.CheerioElement, defaultValue?: string) => void
}

export const Appliers: PaddingApplier = {
   
   // applyPadding adds padding
   applyPadding: ($el: cheerio.CheerioElement, defaultValue: string = "0"): void => {
       const padding = $el.attribs["padding"] || defaultValue
       delete $el.attribs["padding"]
       $el.attribs["style"] = AddToInlineStyle($el.attribs["style"], `padding:${padding}`)
   },
    
   // applyPaddingTop adds padding-top
   applyPaddingTop: ($el: cheerio.CheerioElement, defaultValue: string = "0"): void => {
       const paddingTop = $el.attribs["paddingtop"] || defaultValue
       delete $el.attribs["paddingtop"]
       $el.attribs["style"] = AddToInlineStyle($el.attribs["style"], `padding-top:${paddingTop}`)
   },
   
   // applyPaddingRight adds padding-right
   applyPaddingRight: ($el: cheerio.CheerioElement, defaultValue: string = "0"): void => {
       const paddingRight = $el.attribs["paddingright"] || defaultValue
       delete $el.attribs["paddingright"]
       $el.attribs["style"] = AddToInlineStyle($el.attribs["style"], `padding-right:${paddingRight}`)
   },
   
   // applyPaddingBottom adds padding-bottom
   applyPaddingBottom: ($el: cheerio.CheerioElement, defaultValue: string = "0"): void => {
       const paddingBottom = $el.attribs["paddingbottom"] || defaultValue
       delete $el.attribs["paddingbottom"]
       $el.attribs["style"] = AddToInlineStyle($el.attribs["style"], `padding-bottom:${paddingBottom}`)
   },
   
   // applyPaddingLeft adds padding-left
   applyPaddingLeft: ($el: cheerio.CheerioElement, defaultValue: string = "0"): void => {
       const paddingLeft = $el.attribs["paddingleft"] || defaultValue
       delete $el.attribs["paddingleft"]
       $el.attribs["style"] = AddToInlineStyle($el.attribs["style"], `padding-left:${paddingLeft}`)
   }
}