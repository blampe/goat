(
 ;; GoAT diagrams stored here all end in .txt, so on startup Emacs by default
 ;; will select text-mode for these files.
 ;; Override the user's usual font preference for text-mode with a font known
 ;; to reference standard-dimension glyphs for characters most useful
 ;; in GoAT diagrams.
 ;;
 ;; XX Heavy-handed hack; No help to non-Emacs users.
 (text-mode . (
	       ;; To see the effect of this, issue:
	       ;;    `M-x describe-variable RET face-remapping-alist`
	       (eval . (face-remap-add-relative 'default :family "DejaVu Sans Mono"))

	       ;; Below helps avoid introducing into goat source text any TAB characters,
	       ;; which can be nothing but trouble. 
	       (indent-tabs-mode . nil)
	       )
	    )
 )
