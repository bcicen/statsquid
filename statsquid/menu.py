import curses
from curses import panel

def display_menu(ws,x1,y1,menu_items,attribut1):
    current_option=0
    for o in menu_items:
        ws.addstr(y1,x1,str(o),attribut1[current_option])
        ws.clrtoeol()
        y1+=1
        current_option+=1
    ws.move(2,0)
    ws.border()
    ws.refresh()

def run_menu(menu_items,x=0,y=0,name=None):
    """
    will display the menu at x, y on a newly created window
    then display menu to relative coordinates in that new window called w
    see display_menu above
    """
    max_length = max(len(s) for s in menu_items) + 10
    max_option = len(menu_items)
    current_option = 0
    option_selected = -1
    if name:
        pass
    w = curses.newwin(max_option+2,max_length,y,x)
    w.keypad(1)
    w.refresh()
    while option_selected == -1:
        attribut=[curses.A_NORMAL]*max_option
        attribut[current_option]=curses.A_REVERSE+curses.A_BOLD
        display_menu(w,2,1,menu_items,attribut)
        a = w.getch()
        if a == curses.KEY_DOWN:
            current_option+=1
        elif a == curses.KEY_UP:
            current_option-=1
        elif a==ord('\n') or a == 32 :
        # validation can be done by CR or space bar
            option_selected=current_option
        elif a in range(ord('0'),ord('0')+max_option):
        # in case key pressed is a number
            current_option=a-ord('0')
            option_selected=current_option
        if current_option>max_option-1:
            current_option=max_option-1
        elif current_option <0:
            current_option=0
    return option_selected
