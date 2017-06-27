import {AddToInlineStyle} from '../../util/util'
import cheerio from 'cheerio'

export type MarginApplier = {
    applyMargin: ($el: cheerio.CheerioElement, defaultValue?: string) => void,
    applyMarginTop: ($el: cheerio.CheerioElement, defaultValue?: string) => void,
    applyMarginRight: ($el: cheerio.CheerioElement, defaultValue?: string) => void,
    applyMarginBottom: ($el: cheerio.CheerioElement, defaultValue?: string) => void,
    applyMarginLeft: ($el: cheerio.CheerioElement, defaultValue?: string) => void,
}

export const Appliers: MarginApplier = {
   
   // applyMargin adds margin
   applyMargin: ($el: cheerio.CheerioElement, defaultValue: string = "0"): void => {
       const margin = $el.attribs["margin"] || defaultValue
       delete $el.attribs["margin"]
       $el.attribs["style"] = AddToInlineStyle($el.attribs["style"], `margin:${margin}`)
   },
    
   // applyMarginTop adds margin-top
   applyMarginTop: ($el: cheerio.CheerioElement, defaultValue: string = "0"): void => {
       const marginTop = $el.attribs["margintop"] || defaultValue
       delete $el.attribs["margintop"]
       $el.attribs["style"] = AddToInlineStyle($el.attribs["style"], `margin-top:${marginTop}`)
   },
   
   // applyMarginRight adds margin-right
   applyMarginRight: ($el: cheerio.CheerioElement, defaultValue: string = "0"): void => {
       const marginRight = $el.attribs["marginright"] || defaultValue
       delete $el.attribs["marginright"]
       $el.attribs["style"] = AddToInlineStyle($el.attribs["style"], `margin-right:${marginRight}`)
   },
   
   // applyMarginBottom adds margin-bottom
   applyMarginBottom: ($el: cheerio.CheerioElement, defaultValue: string = "0"): void => {
       const marginBottom = $el.attribs["marginbottom"] || defaultValue
       delete $el.attribs["marginbottom"]
       $el.attribs["style"] = AddToInlineStyle($el.attribs["style"], `margin-bottom:${marginBottom}`)
   },
   
   // applyMarginLeft adds margin-left
   applyMarginLeft: ($el: cheerio.CheerioElement, defaultValue: string = "0"): void => {
       const marginLeft = $el.attribs["marginleft"] || defaultValue
       delete $el.attribs["marginleft"]
       $el.attribs["style"] = AddToInlineStyle($el.attribs["style"], `margin-left:${marginLeft}`)
   }
}