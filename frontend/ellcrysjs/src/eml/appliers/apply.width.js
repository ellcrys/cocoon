import {AddToInlineStyle} from '../../util/util'
import cheerio from 'cheerio'

export type WidthApplier = {
    applyWidth: ($el: cheerio.CheerioElement, defaultValue?: string) => void,
    applyMinWidth: ($el: cheerio.CheerioElement, defaultValue?: string) => void,
    applyMaxWidth: ($el: cheerio.CheerioElement, defaultValue?: string) => void
}

export const Appliers: WidthApplier = {
   
   // applyWidth adds width
   applyWidth: ($el: cheerio.CheerioElement, defaultValue: string = "0"): void => {
       const width = $el.attribs["width"] || defaultValue
       delete $el.attribs["width"]
       $el.attribs["style"] = AddToInlineStyle($el.attribs["style"], `width:${width}`)
   },
   
   // applyMinWidth adds min-width
   applyMinWidth: ($el: cheerio.CheerioElement, defaultValue: string = "0"): void => {
       const minWidth = $el.attribs["minwidth"] || defaultValue
       delete $el.attribs["minwidth"]
       $el.attribs["style"] = AddToInlineStyle($el.attribs["style"], `min-width:${minWidth}`)
   },
   
   // applyMaxWidth adds max-width
   applyMaxWidth: ($el: cheerio.CheerioElement, defaultValue: string = "0"): void => {
       const maxWidth = $el.attribs["maxwidth"] || defaultValue
       delete $el.attribs["maxwidth"]
       $el.attribs["style"] = AddToInlineStyle($el.attribs["style"], `min-width:${maxWidth}`)
   },
}