// @flow
import './flexbox.browser.css'
import cheerio from 'cheerio'
import Appliers from './appliers'


// Browser applies flexbox styles to CML elements
class Browser {

    appliers: typeof Appliers

    constructor() {
        this.appliers = Appliers
    }

    /**
     * Applies flexbox styles to an element
     * @param {cheerio.CheerioElement} $el The element
     * @param {string} attr  The flexbox property to applier
     * @returns {boolean}   true if property was applied or false if not
     * @memberof Browser
     */
    handleFlexboxAttrs($el: cheerio.CheerioElement, attr: string): boolean {
        switch (attr) {
            case 'direction':
                this.appliers.flexbox.applyFlexDirection($el)
                break;
            case 'wrap':
                this.appliers.flexbox.applyFlexWrap($el)
                break;
            case 'justifycontent':
                this.appliers.flexbox.applyJustifyContent($el)
                break;
            case 'alignitems':
                this.appliers.flexbox.applyAlignItems($el)
                break;
            case 'aligncontent':
                this.appliers.flexbox.applyAlignContent($el)
                break;
            case 'order':
                this.appliers.flexbox.applyOrder($el)
                break;
            case 'grow':
                this.appliers.flexbox.applyGrow($el)
                break;
            case 'shrink':
                this.appliers.flexbox.applyShrink($el)
                break;
            case 'basis':
                this.appliers.flexbox.applyBasis($el)
                break;
            case 'alignSelf':
                this.appliers.flexbox.applyAlignSelf($el)
                break;
            default:
                return false
        }
        return true
    }

    /**
     * Applies padding (padding, padding-top, padding-bottom, padding-left, padding-right)
     * @param {cheerio.CheerioElement} $el  The element
     * @param {string} attr                 The padding attribute to apply
     * @returns {boolean}                   true if property was applied or false if not
     * @memberof Browser
     */
    handlePaddingAttrs($el: cheerio.CheerioElement, attr: string): boolean {
        switch (attr) {
            case 'padding': 
                this.appliers.padding.applyPadding($el)
                break;
            case 'paddingtop':
                this.appliers.padding.applyPaddingTop($el)
                break;
            case 'paddingright':
                this.appliers.padding.applyPaddingRight($el)
                break;
            case 'paddingbottom':
                this.appliers.padding.applyPaddingBottom($el)
                break;
            case 'paddingleft':
                this.appliers.padding.applyPaddingLeft($el)
                break;
            default:
                return false
        }
        return true
    }
    
    /**
     * Applies height, min-height and max-height 
     * @param {cheerio.CheerioElement} $el  The element 
     * @param {string} attr                 The attribute
     * @returns {boolean}  true if property was applied
     * @memberof Browser
     */
    handleHeightAttrs($el: cheerio.CheerioElement, attr: string): boolean {
        switch (attr) {
            case 'height': 
                this.appliers.height.applyHeight($el)
                break;
            case 'minheight':
                this.appliers.height.applyMinHeight($el)
                break;
            case 'maxheight':
                this.appliers.height.applyMaxHeight($el)
                break;
            default:
            return false
        }
        return true
    }
    
    /**
     * Applies width, min-width and max-width 
     * @param {cheerio.CheerioElement} $el  The element 
     * @param {string} attr  The attribute
     * @returns {boolean}  true if property was applied
     * @memberof Browser
     */
    handleWidthAttrs($el: cheerio.CheerioElement, attr: string): boolean {
        switch (attr) {
            case 'width': 
                this.appliers.width.applyWidth($el)
                break;
            case 'minwidth':
                this.appliers.width.applyMinWidth($el)
                break;
            case 'maxwidth':
                this.appliers.width.applyMaxWidth($el)
                break;
            default:
            return false
        }
        return true
    }

    /**
     * Applies margin (margin, margin-top, margin-bottom, margin-left, margin-right)
     * @param {cheerio.CheerioElement} $el  The element
     * @param {string} attr                 The margin attribute to apply
     * @returns {boolean}                   true if property was applied or false if not
     * @memberof Browser
     */
    handleMarginAttrs($el: cheerio.CheerioElement, attr: string): boolean {
        switch (attr) {
            case 'margin': 
                this.appliers.margin.applyMargin($el)
                break;
            case 'margintop':
                this.appliers.margin.applyMarginTop($el)
                break;
            case 'marginright':
                this.appliers.margin.applyMarginRight($el)
                break;
            case 'marginbottom':
                this.appliers.margin.applyMarginBottom($el)
                break;
            case 'marginleft':
                this.appliers.margin.applyMarginLeft($el)
                break;
            default:
                return false
        }
        return true
    }

    // apply takes a cheerio element, validates
    // EML attributes and apply corresponding css styles
    apply($el: cheerio.CheerioElement) {
        for (const attr in $el.attribs) {
            var applied = false

            // style (TODO: remove support for arbitrary css styles)
            if (attr === 'style') {
                applied = true
                continue
            }
            
            applied = (!applied) ? this.handleFlexboxAttrs($el, attr) : applied
            applied = (!applied) ? this.handleMarginAttrs($el, attr) : applied
            applied = (!applied) ? this.handlePaddingAttrs($el, attr) : applied
            applied = (!applied) ? this.handleHeightAttrs($el, attr) : applied
            applied = (!applied) ? this.handleWidthAttrs($el, attr) : applied

            if (!applied) throw new Error(`unknown property '${attr}'`);
        }
    }
}

export default Browser